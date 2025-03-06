import { API_BASE_URL } from '@/config';
const initData = window.Telegram?.WebApp?.initData || '';

export async function crashPlace(Amount, CashOutMultiplier) {
    if (!initData) {
        console.error('Telegram WebApp initData is missing');
        return { 
            ok: false, 
            status: 401,
            json: async () => ({ error: 'Ошибка авторизации. Telegram WebApp initData отсутствует.' })
        };
    }

    try {
        const requestBody = { Amount: Amount };

        if (CashOutMultiplier > 1) {
            requestBody.CashOutMultiplier = parseFloat(CashOutMultiplier);
        }

        console.log('Отправка запроса на ставку:', requestBody);
        console.log('URL:', `https://${API_BASE_URL}/games/crashgame/place`);
        
        const response = await fetch(`https://${API_BASE_URL}/games/crashgame/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify(requestBody),
        });

        console.log('Получен ответ со статусом:', response.status);
        
        return response;
    } catch (error) {
        console.error('Ошибка при размещении ставки:', error);
        return { 
            ok: false, 
            status: 500,
            json: async () => ({ error: 'Сетевая ошибка при размещении ставки.' })
        };
    }
}
