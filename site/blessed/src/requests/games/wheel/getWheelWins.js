import {apiClient} from "@/apiClient";

export async function getWheelWins() {
    try {
        const response = await apiClient(`/games/fortunewheel/wins`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

         const data = await response.json();
        
        return await data;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
