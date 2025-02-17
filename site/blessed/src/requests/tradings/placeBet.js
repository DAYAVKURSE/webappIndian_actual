import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function placeBet(Amount, Duration, Direction) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/binary/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({ 
                Amount: Amount,
                Duration: Duration,
                Direction: Direction
            }),
        });

        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
