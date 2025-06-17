import { useState } from "react";
import styles from "./Withdrawal.module.scss";
import { createWithdrawal } from "@/requests";
import { Input } from "@/components";
import { toast } from "react-hot-toast";

export const Withdrawal = () => {
    const [amount, setAmount] = useState(500);
    const [form, setForm] = useState({
        "amount": 0,
        "data": {
            "account_name": "",
            "account_number": "",
            "bank_code": ""
        }
    });
    const [loading, setLoading] = useState(false);

    const validateAccountName = (value) => /^[a-zA-Z\s]{1,30}$/.test(value);
    const validateAccountNumber = (value) => /^[0-9]+$/.test(value);
    const validateBankCode = (value) => /^[A-Z]{4}0[A-Z0-9]{6}$/.test(value);

    const handleFormChange = (e) => {
        const { name, value } = e.target;

        setForm((prevForm) => ({
            ...prevForm,
            data: {
                ...prevForm.data,
                [name]: value,
            },
        }));
    };

    const handleAmountChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) {
            setAmount(Number(value));
        }
    };

    const handleSubmit = async () => {
        const { account_name, account_number, bank_code } = form.data;

        if (!account_name || !account_number || !bank_code || amount <= 0) {
            toast.error('Please fill in all fields and enter a valid amount.');
            return;
        }

        if (!validateAccountName(account_name)) {
            toast.error('Account Name should be up to 30 characters long and only contain letters and spaces.');
            return;
        }
        if (!validateAccountNumber(account_number)) {
            toast.error('Account Number should only contain digits.');
            return;
        }
        if (!validateBankCode(bank_code)) {
            toast.error('Bank Code (IFSC) must be exactly 11 characters: the format should be "AAAA0AAAAAA".');
            return;
        }

        
        toast.success('Your withdraw application has been created! Please contact our support team to confirm your application.');

        // setLoading(true);
        // try {
        //     const response = await createWithdrawal(
        //         amount,
        //         account_name,
        //         account_number,
        //         bank_code
        //     );

        //     if (response.status === 200) {
        //         toast.success('Withdrawal created successfully.');

        //         setForm({
        //             "amount": 0,
        //             "data": {
        //                 "account_name": "",
        //                 "account_number": "",
        //                 "bank_code": ""
        //             }
        //         });
        //         setAmount(0);
        //     } else {
        //         toast.error(response.message || 'Failed to create withdrawal. Please try again.');
        //     }
        // } catch (error) {
        //     console.log('Error creating withdrawal:', error);
        // } finally {
        //     setLoading(false);
        // }
    };

    return (
        <div className={styles.withdrawal}>
            <div className={styles.withdrawal__withdrawal}>
                <h1 className={styles.withdrawal__withdrawal_title}>Withdrawal</h1>
                <div className={styles.withdrawal__withdrawal__balance}>
                    <p className={styles.withdrawal__withdrawal__balance_text}>Available for withdrawal</p>
                </div>
                <div className={styles.withdrawal__amount}>
                    <div className={styles.withdrawal__amount__input}>
                        <div className={styles.withdrawal__amount__input_container}>
                            
                            <input
                                type="text"
                                placeholder="500"
                                value={amount}
                                onChange={handleAmountChange}
                            />
                        </div>
                    
                    </div>
                </div>
                <div className={styles.withdrawal__form}>
                    <Input
                        type="text"
                        placeholder="Account Name"
                        name="account_name"
                        value={form.data.account_name}
                        onChange={handleFormChange}
                    />
                    <Input
                        type="text"
                        placeholder="Account Number"
                        name="account_number"
                        value={form.data.account_number}
                        onChange={handleFormChange}
                    />
                    <Input
                        type="text"
                        placeholder="Bank Code"
                        name="bank_code"
                        value={form.data.bank_code}
                        onChange={handleFormChange}
                    />
                    {/* <Button
                        label={loading ? "Processing..." : "Withdrawal"}
                        
                        disabled={loading}
                        color="lightYellow"
                    /> */}
                    <button  onClick={handleSubmit} className={styles.widthdraw}>{loading ? "Processing..." : "Withdrawal"}</button>
                </div>
            </div>
        </div>
    );
};
