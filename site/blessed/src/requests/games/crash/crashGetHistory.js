import {apiClient} from "@/apiClient";


export async function crashGetHistory() {
    try {
        const response = await apiClient(`/games/crashgame/history`, {
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