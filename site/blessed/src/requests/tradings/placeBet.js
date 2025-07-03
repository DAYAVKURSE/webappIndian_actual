import {apiClient} from "@/apiClient";

export async function placeBet(Amount, Duration, Direction) {
    try {
        const response = await apiClient(`/games/binary/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                Amount: Amount,
                Duration: Duration,
                Direction: Direction
            }),
        });

        const data = await response.json();

        return await data;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
