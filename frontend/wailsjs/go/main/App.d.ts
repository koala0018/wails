export interface FileSelection {
  path: string;
}

export interface SplitRequest {
  inputPath: string;
  outputDir: string;
  segmentLengthSec: number;
}

export interface SplitResult {
  outputDir: string;
  segmentCount: number;
  generatedFiles: string[];
  command: string;
}

export interface EnvironmentStatus {
  ffmpegReady: boolean;
  ffmpegPath: string;
  message: string;
}

export function GetEnvironmentStatus(): Promise<EnvironmentStatus>;
export function SelectOutputDirectory(): Promise<FileSelection>;
export function SelectVideoFile(): Promise<FileSelection>;
export function SplitVideo(arg1: SplitRequest): Promise<SplitResult>;
