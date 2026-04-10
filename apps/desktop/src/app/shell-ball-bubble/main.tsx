import ReactDOM from "react-dom/client";
import { AppProviders } from "@/features/shared/AppProviders";
import { ShellBallBubbleWindow } from "@/features/shell-ball/ShellBallBubbleWindow";
import "@/features/shell-ball/shellBall.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <AppProviders>
    <ShellBallBubbleWindow />
  </AppProviders>,
);
