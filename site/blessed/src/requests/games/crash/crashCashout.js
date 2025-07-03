import {apiClient} from "@/apiClient";

export async function crashCashout() {
  try {
    const response = await apiClient(`/games/crashgame/cashout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

      const data = await response.json();

    return data;
  } catch (error) {
    console.error('Error when withdrawing funds:', error);
    return {
      ok: false,
      status: 500,
      json: async () => ({ error: 'Network error when withdrawing funds.' }),
    };
  }
}
