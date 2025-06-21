import {apiClient} from "@/apiClient";

export async function getLeaders(type) {
    try {
        const response = await apiClient(`/leaders/get?period=${type}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        const contentLength = await response.json();
        
        return contentLength;
    } catch (error) {
        console.error('Error registering user:', error);
    }
}
