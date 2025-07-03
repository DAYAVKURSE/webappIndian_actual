import { apiClient } from '@/apiClient';

export async function rollDice(Amount, WinPercent, Direction) {
  try {
    const response = await apiClient(`/games/dice/place`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        Amount: Amount,
        WinPercent: WinPercent,
        Direction: Direction,
      }),
    });

    const data = await response.json();

    return await data;
  } catch (error) {
    console.error('Error registering user:', error);
    throw error;
  }
}
