import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function getLeaders(type) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/leaders/get?period=${type}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            }
        });

        const contentLength = await response.json();
        
        return contentLength;
    } catch (error) {
        console.error('Error registering user:', error);
    }
}
