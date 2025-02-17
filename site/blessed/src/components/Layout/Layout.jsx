import toastStyles from "@/scss/toast.module.scss";
import { Toaster, useToasterStore, toast } from 'react-hot-toast';
import { Header, Footer } from "@/components";
import { Outlet } from "react-router-dom";
import { getMe } from "@/requests";
import { useEffect } from "react";

const TOAST_LIMIT = 3;

export const Layout = () => {
  const { toasts } = useToasterStore();

  useEffect(() => {
    getMe();
  }, []);

  useEffect(() => {
    toasts
      .filter((t) => t.visible)
      .filter((_, i) => i >= TOAST_LIMIT)
      .forEach((t) => toast.dismiss(t.id));
  }, [toasts]);

  return (
    <>
      <div>
        <Toaster
          toastOptions={{
            position: 'top-center',
            success: {
              className: toastStyles.toastSuccess,
              iconTheme: {
                primary: '#EDFF8C',
                secondary: '#0B0B0B',
              },
            },
            error: {
              className: toastStyles.toastError,
              iconTheme: {
                primary: '#FFC397',
                secondary: '#0B0B0B',
              },
            },
            default: {
              className: toastStyles.toast,
            },
          }}
        />
      </div>
      <Header />
      <main><Outlet /></main>
      <Footer />
    </>
  );
};
