import {apiClient} from "@/apiClient";

export async function crashPlace(Amount, CashOutMultiplier) {


    try {
        const requestBody = { amount: Number(Amount) };

        if (CashOutMultiplier > 1) {
            requestBody.CashOutMultiplier = parseFloat(CashOutMultiplier);
        }

        
        const response = await apiClient(`/games/crashgame/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
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
