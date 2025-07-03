import {apiClient} from "@/apiClient";

export async function roulettePlaceBet(Amount, Color) {
    try {
        const response = await apiClient(`/games/roulettex14/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                Amount: Amount,
                Color: Color,
            }),
        });

        const data = await response.json();

        return await data;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
