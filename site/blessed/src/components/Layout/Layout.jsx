import toastStyles from "@/scss/toast.module.scss";
import { Toaster, useToasterStore, toast } from "react-hot-toast";
import { Header, Footer } from "@/components";
import { Outlet,
  //  useNavigate
   } from "react-router-dom";
import { useEffect } from "react";
// import { getMe } from "../../requests/users/getMe";

const TOAST_LIMIT = 3;

export const Layout = () => {
  const { toasts } = useToasterStore();
  // const navigate = useNavigate();

  // Получение данных пользователя (аналог приватного роутера)
  // useEffect(() => {
  //   getMe().catch((error) => {
  //     console.error("User not authenticated:", error);
  //     navigate("/onboarding");
  //   });
  // }, []);

  // Ограничение количества видимых тостов
  useEffect(() => {
    toasts
      .filter((t) => t.visible)
      .filter((_, i) => i >= TOAST_LIMIT)
      .forEach((t) => toast.dismiss(t.id));
  }, [toasts]);

  return (
    <>
      <Toaster
        toastOptions={{
          position: "top-center",
          success: {
            className: toastStyles.toastSuccess,
            style: { border: "1px solid #565656" },
            iconTheme: {
              primary: "#0D99FF",
              secondary: "#1A1B20",
            },
          },
          error: {
            className: toastStyles.toastError,
            style: { border: "1px solid #565656" },
            iconTheme: {
              primary: "#C81100",
              secondary: "#1A1B20",
            },
          },
          default: {
            className: toastStyles.toast,
            style: { border: "1px solid #565656" },
          },
        }}
      />

      <Header />
      <main>
        <Outlet />
      </main>
      <Footer />
    </>
  );
};
