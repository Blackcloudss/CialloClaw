import { ShellBallDevLayer } from "./ShellBallDevLayer";
import { shouldShowShellBallDemoSwitcher } from "./shellBall.dev";
import { ShellBallSurface } from "./ShellBallSurface";
import { useShellBallInteraction } from "./useShellBallInteraction";
import { getShellBallMotionConfig } from "./shellBall.motion";
import { useShellBallCoordinator } from "./useShellBallCoordinator";

type ShellBallAppProps = {
  isDev?: boolean;
};

export function ShellBallApp({ isDev = false }: ShellBallAppProps) {
  const {
    visualState,
    inputValue,
    voicePreview,
    handlePrimaryClick,
    handleRegionEnter,
    handleRegionLeave,
    handlePressStart,
    handlePressMove,
    handlePressEnd,
    handleSubmitText,
    handleAttachFile,
    handleInputFocusChange,
    setInputValue,
    handleForceState,
  } = useShellBallInteraction();
  const motionConfig = getShellBallMotionConfig(visualState);
  const showDemoSwitcher = shouldShowShellBallDemoSwitcher(isDev);

  useShellBallCoordinator({
    visualState,
    inputValue,
    voicePreview,
    setInputValue,
    onRegionEnter: handleRegionEnter,
    onRegionLeave: handleRegionLeave,
    onInputFocusChange: handleInputFocusChange,
    onSubmitText: handleSubmitText,
    onAttachFile: handleAttachFile,
    onPrimaryClick: handlePrimaryClick,
  });

  return (
    <ShellBallSurface
      visualState={visualState}
      voicePreview={voicePreview}
      motionConfig={motionConfig}
      onPrimaryClick={handlePrimaryClick}
      onRegionEnter={handleRegionEnter}
      onRegionLeave={handleRegionLeave}
      onPressStart={handlePressStart}
      onPressMove={handlePressMove}
      onPressEnd={handlePressEnd}
    >
      {showDemoSwitcher ? (
        <ShellBallDevLayer value={visualState} onChange={handleForceState} />
      ) : null}
    </ShellBallSurface>
  );
}
