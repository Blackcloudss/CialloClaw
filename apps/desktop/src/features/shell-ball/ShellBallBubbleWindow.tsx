import type { ShellBallVisualState } from "./shellBall.types";
import { useShellBallHelperWindowSnapshot } from "./useShellBallCoordinator";
import { ShellBallBubbleZone } from "./components/ShellBallBubbleZone";

type ShellBallBubbleWindowProps = {
  visualState?: ShellBallVisualState;
};

export function ShellBallBubbleWindow({ visualState }: ShellBallBubbleWindowProps) {
  const snapshot = useShellBallHelperWindowSnapshot({ role: "bubble" });
  const resolvedVisualState = visualState ?? snapshot.visualState;

  return (
    <div className="shell-ball-window shell-ball-window--bubble" aria-label="Shell-ball bubble window">
      <ShellBallBubbleZone visualState={resolvedVisualState} />
    </div>
  );
}
