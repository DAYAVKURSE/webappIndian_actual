import { API_BASE_URL } from '@/config';
import { toast } from "react-hot-toast"; // Добавим импорт toast чтоб потом уведы оформить


export async function createWithdrawal(amount, accountName, accountNumber, bankCode) {
    const body = {
        amount: amount,
        payment_system: "imps",
        data: {
            account_name: accountName,
            account_number: accountNumber,
            bank_code: bankCode
        }
    };

    try {
        const response = await fetch(`${API_BASE_URL}/payments/withdrawal`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
        });

        if (response.status === 404 || response.status === 406 || response.status === 400 || response.status === 402) {
            const data = await response.json();
            const message = data.error.charAt(0).toUpperCase() + data.error.slice(1);
            return { status: response.status, message };
        }

        const data = await response.json();
        // Добавим уведомление об успешной обработке
        if (response.status === 200) {
            toast.success('Your withdrawal request is being processed');
        }
        return { status: response.status, data };
    } catch (error) {
        console.error('Error creating withdrawal:', error);
        throw error;
    }
}