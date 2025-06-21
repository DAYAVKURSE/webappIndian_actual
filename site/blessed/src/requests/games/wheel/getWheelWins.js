import {apiClient} from "@/apiClient";

export async function getWheelWins() {
    try {
        const response = await apiClient(`/games/fortunewheel/wins`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        
        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
