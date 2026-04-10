import ReactDOM from "react-dom/client";
import { AppProviders } from "@/features/shared/AppProviders";
import { ShellBallInputWindow } from "@/features/shell-ball/ShellBallInputWindow";
import "@/features/shell-ball/shellBall.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <AppProviders>
    <ShellBallInputWindow />
  </AppProviders>,
);
