import * as THREE from 'three'
import type { RenderMode, LightingPreset, CameraPreset } from './threeScene';
import type { BaseModelCandidate, ViewerPayload } from '#lib/bindings/changeme/modpreview/models.ts';


export interface BuiltMesh {
  mesh: THREE.Mesh;
  name: string;
  component: string | null;
  condition: string | null;
  vertexCount: number;
  triangleCount: number;
}

export interface RetextureInfo {
  isRetexture: boolean;
  baseModelSource: string | null;
  candidates: BaseModelCandidate[];
}

export interface ModViewerModalProps {
  modId: string;
  packId: string;
  onClose: () => void;
  /** When true, suppresses the fixed inset-0 overlay wrapper so the viewer fills a parent container */
  embedded?: boolean;
  /** Callback to use the captured 3D view snapshot as the mod preview image */
  onUseAsPreview?: (imageSrc: string) => void;
  /** Optional pre-loaded payload (e.g. for GameBanana in-flight 3D preview) */
  initialPayload?: ViewerPayload;
  /** Whether this viewer session is an in-flight GameBanana mod preview */
  isGbPreview?: boolean;
  /** Callback invoked when the user clicks 'Install Mod to Library' in GameBanana preview mode */
  onInstallGbMod?: () => Promise<void> | void;
}

export type {
  RenderMode,
  LightingPreset,
  CameraPreset,
  BaseModelCandidate,
};