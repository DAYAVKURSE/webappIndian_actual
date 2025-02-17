import { create } from 'zustand';

const saveToLocalStorage = (state) => {
  localStorage.setItem('store', JSON.stringify(state));
};

const loadFromLocalStorage = () => {
  const storedState = localStorage.getItem('store');
  return storedState ? JSON.parse(storedState) : {};
};

const useStore = create((set) => ({
    ...{
      userName: 'username',
      BalanceRupee: 0,
      dailyClicks: 0,
      BiPerClick: 0,
      BalanceBi: 0,
      referredBy: null,
    },
    ...loadFromLocalStorage(),

    setUserName: (newName) => set((state) => {
        const newState = { ...state, userName: newName };
        saveToLocalStorage(newState);
        return newState;
    }),

    setBalanceRupee: (newBalanceRupee) => set((state) => {
        const newState = { ...state, BalanceRupee: newBalanceRupee };
        saveToLocalStorage(newState);
        return newState;
    }),
    
    increaseBalanceRupee: (amount) => set((state) => {
        const newState = { ...state, BalanceRupee: state.BalanceRupee + amount };
        saveToLocalStorage(newState);
        return newState;
    }),
    
    decreaseBalanceRupee: (amount) => set((state) => {
        const newState = { ...state, BalanceRupee: state.BalanceRupee - amount };
        saveToLocalStorage(newState);
        return newState;
    }),
    
    setDailyClicks: (newDailyClicks) => set((state) => {
        const newState = { ...state, dailyClicks: newDailyClicks };
        saveToLocalStorage(newState);
        return newState;
    }),

    increaseBalanceBi: () => set((state) => {
        const newState = { BalanceBi: state.BalanceBi + state.BiPerClick };
        saveToLocalStorage(newState);
        return newState;
    }),

    decreaseBalanceBi: (amount) => set((state) => {
        const newState = { BalanceBi: state.BalanceBi - amount };
        saveToLocalStorage(newState);
        return newState;
    }),

    incrementDailyClicks: () => set((state) => {
        const newState = { dailyClicks: state.dailyClicks + 1 };
        saveToLocalStorage(newState);
        return newState;
    }),

    addBalanceBi: (newBalanceBi) => set((state) => {
        const newState = { BalanceBi: state.BalanceBi + newBalanceBi };
        saveToLocalStorage(newState);
        return newState;
    }),

    setBalanceBi: (newBalanceBi) => set((state) => {
        const newState = { BalanceBi: newBalanceBi };
        saveToLocalStorage(newState);
        return newState;
    }),

    setReferredBy: (newReferredBy) => set((state) => {
        const newState = { referredBy: newReferredBy };
        saveToLocalStorage(newState);
        return newState;
    }),
}));

export default useStore;
