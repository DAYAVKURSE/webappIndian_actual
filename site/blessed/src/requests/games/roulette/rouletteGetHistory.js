import {apiClient} from "@/apiClient";


export async function rouletteGetHistory() {
    try {
        const response = await apiClient(`/games/roulettex14/history`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        return await response.json();
    } catch (error) {
        console.error('Error fetching game history:', error);
        throw error;
    }
}