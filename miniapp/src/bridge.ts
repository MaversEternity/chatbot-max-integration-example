// Thin wrapper around MAX Bridge (window.WebApp)
// In MAX messenger, WebApp is injected by https://st.max.ru/js/max-web-app.js
// In browser, it may be undefined — all calls are safe.

const webapp = (window as any).WebApp;

export const bridge = {
  get user() {
    return webapp?.initDataUnsafe?.user ?? { first_name: 'Иван', last_name: 'Петрович' };
  },
  get platform(): string {
    return webapp?.platform ?? 'web';
  },
  hapticImpact(style: 'light' | 'medium' | 'heavy' = 'light') {
    webapp?.HapticFeedback?.impactOccurred(style);
  },
  hapticNotification(type: 'success' | 'error' | 'warning') {
    webapp?.HapticFeedback?.notificationOccurred(type);
  },
  hapticSelection() {
    webapp?.HapticFeedback?.selectionChanged();
  },
  showBackButton() {
    webapp?.BackButton?.show();
  },
  hideBackButton() {
    webapp?.BackButton?.hide();
  },
  onBackButton(cb: () => void) {
    webapp?.BackButton?.onClick(cb);
  },
  offBackButton(cb: () => void) {
    webapp?.BackButton?.offClick(cb);
  },
  close() {
    webapp?.close();
  },
};
