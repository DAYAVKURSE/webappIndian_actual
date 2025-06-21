import {apiClient} from "@/apiClient";

export async function exchange(AmountBcoins) {
    try {
        const response = await apiClient(`/users/exchange`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                AmountBcoins: AmountBcoins
            }),
        });
        
        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
