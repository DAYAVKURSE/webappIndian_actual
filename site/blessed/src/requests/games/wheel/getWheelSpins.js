import {apiClient} from "@/apiClient";
export async function getWheelSpins() {
    try {
        const response = await apiClient(`/games/fortunewheel/spins`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
