<<<<<<< HEAD
import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
=======
import { API_BASE_URL } from '@/config';
>>>>>>> main
const initData = window.Telegram?.WebApp?.initData || '';

export async function crashCashout() {
    if (!initData) {
        console.error('Telegram WebApp initData is missing');
        return { 
            ok: false, 
            status: 401,
            json: async () => ({ error: 'Ошибка авторизации. Telegram WebApp initData отсутствует.' })
        };
    }

    try {
        console.log('Отправка запроса на вывод средств');
        console.log('URL:', `https://${API_BASE_URL}/games/crashgame/cashout`);
        
        const response = await fetch(`https://${API_BASE_URL}/games/crashgame/cashout`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        console.log('Получен ответ со статусом:', response.status);
        
        return response;
    } catch (error) {
        console.error('Ошибка при выводе средств:', error);
        return { 
            ok: false, 
            status: 500,
            json: async () => ({ error: 'Сетевая ошибка при выводе средств.' })
        };
    }
}
