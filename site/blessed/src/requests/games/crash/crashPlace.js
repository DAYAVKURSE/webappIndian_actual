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

        
        return response;
    } catch (error) {
        console.error('Error when placing a bid:', error);
        return { 
            ok: false, 
            status: 500,
            json: async () => ({ error: 'Network error when placing a bet.' })
        };
    }
}
