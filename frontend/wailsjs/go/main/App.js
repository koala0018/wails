// @ts-check
// This file is manually aligned with the current Go bindings.

export function GetEnvironmentStatus() {
  return window['go']['main']['App']['GetEnvironmentStatus']();
}

export function SelectOutputDirectory() {
  return window['go']['main']['App']['SelectOutputDirectory']();
}

export function SelectVideoFile() {
  return window['go']['main']['App']['SelectVideoFile']();
}

export function SplitVideo(arg1) {
  return window['go']['main']['App']['SplitVideo'](arg1);
}
