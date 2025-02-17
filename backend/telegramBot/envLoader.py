import os

def load_env():
    env_file = ".env"
    try:
        with open(env_file) as file:
            for line in file:
                if line.strip() and not line.startswith('#'):
                    key, value = line.strip().split('=', 1)
                    os.environ[key] = value
    except:
        pass
    botToken = os.environ.get('TOKEN')
    webhook_port = int(os.environ.get("WEBHOOK_PORT", 8443))
    webhook_host = os.environ.get("WEBHOOK_HOST")
    return botToken, webhook_host, webhook_port