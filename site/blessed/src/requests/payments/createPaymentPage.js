import { API_BASE_URL } from '@/config';
import { toast } from "react-hot-toast";

const initData = window.Telegram.WebApp.initData;

export async function createPaymentPage(amount) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/payments/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({
                amount: amount,
                payment_systems: ["imps", "neft", "rtgs", "upi"]
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            if (response.status === 406) {
                toast.error('Minimum deposit amount is 500 rupees');
            } else {
                const message = data.error ? data.error.charAt(0).toUpperCase() + data.error.slice(1) : 'Error creating payment page';
                toast.error(message);
            }
            return null;
        }

        return data;
    } catch (error) {
        console.error('Error creating payment page:', error);
        toast.error('Error creating payment page');
        throw error;
    }
}