import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function rollDice(Amount, WinPercent, Direction) {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/games/dice/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({ 
                Amount: Amount,
                WinPercent: WinPercent,
                Direction: Direction
            }),
        });

        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
