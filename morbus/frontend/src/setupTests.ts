import '@testing-library/jest-dom';

(window as any).runtime = {
  EventsOnMultiple: () => () => {},
  EventsEmit: () => {},
};
