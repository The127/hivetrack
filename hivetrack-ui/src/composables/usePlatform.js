const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;

export function usePlatform() {
  const modKey = isMac ? "⌘" : "Ctrl";
  return { isMac, modKey };
}
