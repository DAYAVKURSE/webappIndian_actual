import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function spinWheel() {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/fortunewheel/spin`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        const data = await response.json();
        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
