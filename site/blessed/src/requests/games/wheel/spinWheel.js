import {apiClient} from "@/apiClient";

export async function spinWheel() {
    try {
        const response = await apiClient(`/games/fortunewheel/spin`, {
            method: 'POST',
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
