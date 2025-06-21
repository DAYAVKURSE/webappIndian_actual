

import { Layout } from '@/components';
import { PrivateRoute } from './PrivateRoute';

import {
  Home, Terms, Loading, Onboarding, Clicker, Trading, Other,
  Wallet, Profile, Games, Dice, Nvuti, Pass, FortuneWheel,
  Howtoplay, Faq, Roulette, Topup, Withdrawal, Crash, Exchange
} from '@/pages';
import { Login } from '../pages/Login/Login';

export const appRoutes = [
  { path: '/login', element: <Login /> },

  {
    path: '/',
    element: <PrivateRoute />, 
    children: [
      { index: true, element: <Loading /> },
      { path: 'landing', element: <Home /> },
      { path: 'terms', element: <Terms /> },
      { path: 'onboarding', element: <Onboarding /> },
      {
        path: '/',
        element: <Layout />,
        children: [
          { path: 'clicker', element: <Clicker /> },
          { path: 'trading', element: <Trading /> },
          { path: 'other', element: <Other /> },
          { path: 'other/how-to-play', element: <Howtoplay /> },
          { path: 'other/faq', element: <Faq /> },
          { path: 'profile', element: <Profile /> },
          { path: 'wallet', element: <Wallet /> },
          { path: 'wallet/topup', element: <Topup /> },
          { path: 'wallet/withdrawal', element: <Withdrawal /> },
          { path: 'wallet/exchange', element: <Exchange /> },
          { path: 'games', element: <Games /> },
          { path: 'wheel', element: <FortuneWheel /> },
          { path: 'pass', element: <Pass /> },
          { path: 'games/dice', element: <Dice /> },
          { path: 'games/nvuti', element: <Nvuti /> },
          { path: 'games/roulette', element: <Roulette /> },
          { path: 'games/crash', element: <Crash /> },
        ],
      },
    ],
  },
];
