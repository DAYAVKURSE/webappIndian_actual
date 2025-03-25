import { API_BASE_URL } from '@/config';
import { toast } from "react-hot-toast";

const initData = window.Telegram.WebApp.initData;

export async function createPaymentPage(amount) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/api/payments/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({
                amount: amount
            }),
        });

        if (response.status === 404 || response.status === 406 || response.status === 400) {
            const data = await response.json();
            const message = data.error.charAt(0).toUpperCase() + data.error.slice(1);
            toast.error(message);
            return;
        }

        if (response.status === 406) {
            toast.error('The minimum top-up amount is 500 rupees. Please enter an amount equal to or greater than this.');
            return;
        }

        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error creating payment page:', error);
        toast.error('Error creating payment page');
        throw error;
    }
}