import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { signUp } from '../../requests/users/auth/signup';
import { login } from '../../requests';

export const useAuthForm = (mode) => {
  const [form, setForm] = useState({
    username: '',
    password: '',
    confirmPassword: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    setError('');
  }, [mode]);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
    setError('');
  };

  const handleSubmit = async (e) => {
  e.preventDefault();
  setLoading(true);
  setError('');

  try {
    if (mode === 'register') {
      if (form.password !== form.confirmPassword) {
        setError('Пароли не совпадают');
        setLoading(false);
        return;
      }

      const result = await signUp({
        nickname: form.username,
        password: form.password,
      });

      // Проверяем статус или success-поле из ответа сервера
      if (result?.status === 'ok') {
        navigate(0); // перезагружаем страницу
      } else {
        throw new Error('Registration failed');
      }

    } else {
      const result = await login({
        nickname: form.username,
        password: form.password,
      });

      if (result?.status === 'ok') {
        navigate('/');
      } else {
        throw new Error('Login failed');
      }
    }
  } catch (err) {
    console.error(err);
    setError('Ошибка. Проверьте данные и попробуйте снова.');
  } finally {
    setLoading(false);
  }
};

  return {
    form,
    handleChange,
    handleSubmit,
    error,
    loading,
  };
};
