import {apiClient} from "@/apiClient";

export async function crashCashout() {
  try {
    const response = await apiClient(`/games/crashgame/cashout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    return response;
  } catch (error) {
    console.error('Ошибка при выводе средств:', error);
    return {
      ok: false,
      status: 500,
      json: async () => ({ error: 'Сетевая ошибка при выводе средств.' }),
    };
  }
}
