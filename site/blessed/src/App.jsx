import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { Layout } from "@/components";
import { Home, Terms, Loading, Onboarding, Clicker, Trading, Other, Wallet, Profile, Games, Dice, Nvuti, Pass, FortuneWheel, Howtoplay, Faq, Roulette, Topup, Withdrawal, Crash, Exchange } from "@/pages";

export default function App() {
  return (
    <Router>
      <Routes>
        <Route index element={<Loading />} exact />
        
        <Route path="/landing" element={<Home />} exact />
        <Route path="/terms" element={<Terms />} exact />

        <Route path="/onboarding" element={<Onboarding />} exact />
        <Route path="/" element={<Layout />}>
          <Route path="clicker" element={<Clicker />} exact />
          <Route path="trading" element={<Trading />} exact />

          <Route path="other" element={<Other />} exact />
          <Route path="other/how-to-play" element={<Howtoplay />} exact />
          <Route path="other/faq" element={<Faq />} exact />

          <Route path="profile" element={<Profile />} exact />
          <Route path="wallet" element={<Wallet />} exact />
          <Route path="wallet/topup" element={<Topup />} exact />
          <Route path="wallet/withdrawal" element={<Withdrawal />} exact />
          <Route path="wallet/exchange" element={<Exchange />} exact />

          <Route path="games/" element={<Games />} exact />
          <Route path="wheel" element={<FortuneWheel />} exact />
          <Route path="pass" element={<Pass />} exact />
          <Route path="games/dice" element={<Dice />} exact />
          <Route path="games/nvuti" element={<Nvuti />} exact />
          <Route path="games/roulette" element={<Roulette />} exact />
          <Route path="games/crash" element={<Crash />} exact />
        </Route>
      </Routes>
    </Router>
  )
};
