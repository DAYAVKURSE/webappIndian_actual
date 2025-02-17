import logging
from aiogram import Bot, Dispatcher, types
from aiogram.filters import Command
from aiogram.types import Message, InlineKeyboardButton, InlineKeyboardMarkup, WebAppInfo, BotCommand
from aiogram.webhook.aiohttp_server import SimpleRequestHandler, setup_application
from aiohttp import web
from envLoader import load_env
import ssl
from collections import OrderedDict
import time
from typing import Dict, Tuple, Any, Optional
from threading import Lock

# Configuration
REQUIRE_CHANNEL_SUB = False  # Set this to False to make channel subscription optional
channel_username = "@BiTRaveofficial"

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Load environment variables
token, webhook_host, webhook_port = load_env()
webapp_url = "https://bitrave.co/"
webhook_url = f"https://{webhook_host}:{webhook_port}"

# Initialize bot and dispatcher
bot = Bot(token)
dp = Dispatcher()

class TimedCache:
    def __init__(self, max_age_seconds: int = 3600, max_size: int = 10000):
        self._max_age = max_age_seconds
        self._max_size = max_size
        self._cache: OrderedDict[int, Tuple[Any, float]] = OrderedDict()
        self._lock = Lock()

    def __setitem__(self, key: int, value: Any) -> None:
        with self._lock:
            self._cleanup()
            self._cache[key] = (value, time.monotonic())
            if len(self._cache) > self._max_size:
                self._cache.popitem(last=False)

    def __getitem__(self, key: int) -> Any:
        with self._lock:
            item = self._cache.get(key)
            if item is None:
                raise KeyError(key)
            value, timestamp = item
            if time.monotonic() - timestamp > self._max_age:
                del self._cache[key]
                raise KeyError(key)
            self._cache.move_to_end(key)
            return value

    def get(self, key: int, default: Any = None) -> Any:
        try:
            return self[key]
        except KeyError:
            return default

    def _cleanup(self) -> None:
        current_time = time.monotonic()
        while self._cache and current_time - next(iter(self._cache.values()))[1] > self._max_age:
            self._cache.popitem(last=False)

    def __len__(self) -> int:
        with self._lock:
            return len(self._cache)

    def __contains__(self, key: int) -> bool:
        with self._lock:
            return key in self._cache

# Initialize the TimedCache with a 1-hour expiration and max size of 10,000
referral_codes = TimedCache(max_age_seconds=600, max_size=10000)

# Function to check if user is subscribed to the channel
async def is_subscribed(user_id: int) -> bool:
    if not REQUIRE_CHANNEL_SUB:
        return True
    try:
        member = await bot.get_chat_member(chat_id=channel_username, user_id=user_id)
        return member.status not in ("left", "kicked")
    except Exception as e:
        return False

# Create keyboard markup
def get_keyboard(is_member: bool, referral_code: str = "") -> InlineKeyboardMarkup:
    # Add referral code to webapp URL if available
    webapp_url_with_referral = f"{webapp_url}?referral={referral_code}" if referral_code else webapp_url
    
    if is_member or not REQUIRE_CHANNEL_SUB:
        keyboard = [[InlineKeyboardButton(text="Open App", web_app=WebAppInfo(url=webapp_url_with_referral))]]
    else:
        keyboard = [
            [InlineKeyboardButton(text="Subscribe to Channel", url=f"https://t.me/{channel_username.lstrip('@')}")],
            [InlineKeyboardButton(text="I've subscribed!", callback_data="check_sub")]
        ]
    return InlineKeyboardMarkup(inline_keyboard=keyboard)

def store_referral_code(user_id: int, referral_code: str) -> None:
    referral_codes[user_id] = referral_code

@dp.message(Command('start'))
async def start_command(message: Message) -> None:
    user_id = message.from_user.id
    
    args = message.text.split()
    referral_code = args[1] if len(args) > 1 else ""
    
    if referral_code:
        store_referral_code(user_id, referral_code)
    
    is_member = True
    if REQUIRE_CHANNEL_SUB:
        try:
            is_member = await is_subscribed(user_id)
        except Exception as e:
            is_member = False
    
    if is_member or not REQUIRE_CHANNEL_SUB:
        text = "Welcome! Thanks for joining in!"
    else:
        text = "Please subscribe to our channel to use the WebApp."
    
    await message.reply(text, reply_markup=get_keyboard(is_member, referral_code))

@dp.callback_query(lambda c: c.data == 'check_sub')
async def check_subscription(callback_query: types.CallbackQuery) -> None:
    if not REQUIRE_CHANNEL_SUB:
        referral_code = referral_codes.get(callback_query.from_user.id, "")
        await callback_query.message.edit_text(
            "Welcome! Thanks for joining in!",
            reply_markup=get_keyboard(True, referral_code)
        )
        return

    user_id = callback_query.from_user.id
    
    try:
        is_member = await is_subscribed(user_id)
    except Exception as e:
        await callback_query.answer("An error occurred. Please try again later.")
        return
    
    if is_member:
        referral_code = referral_codes.get(user_id, "")
        await callback_query.message.edit_text(
            "Great! You're now subscribed. You can use the WebApp.",
            reply_markup=get_keyboard(True, referral_code)
        )
    else:
        await callback_query.answer("You're not subscribed yet. Please subscribe to the channel.")

async def set_commands(bot: Bot):
    commands = [
        BotCommand(command="/start", description="Start the bot")
    ]
    await bot.set_my_commands(commands)
    logger.info("Bot commands have been set")

async def on_startup(app: web.Application):
    await bot.set_webhook(f"{webhook_url}/webhook")
    await set_commands(bot)
    logger.info("Bot started and webhook set")

def main():
    # Create SSL context
    context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    context.load_cert_chain('./ssl/certificate.crt', './ssl/private.key')

    # Set up the application
    app = web.Application()
    SimpleRequestHandler(dispatcher=dp, bot=bot).register(app, path="/webhook")
    setup_application(app, dp, bot=bot)

    # Set up startup handler
    app.on_startup.append(on_startup)

    logger.info(f"Starting the bot... Channel subscription requirement: {'ON' if REQUIRE_CHANNEL_SUB else 'OFF'}")
    # Start the web application
    web.run_app(app, host="0.0.0.0", port=webhook_port, ssl_context=context)

if __name__ == '__main__':
    main()