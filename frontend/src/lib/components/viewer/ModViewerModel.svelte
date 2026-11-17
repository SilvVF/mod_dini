<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { Layers, Sliders } from "@lucide/svelte";
  import type { ModSaved, Pack } from "#lib/bindings/changeme/models.ts";
  import type {
    ViewerPayload,
    ViewerToggle,
  } from "#lib/bindings/changeme/modpreview/models.ts";
  import { createThreeScene, type ThreeSceneAPI } from "./threeScene";
  import type {
    BuiltMesh,
    CameraPreset,
    LightingPreset,
    RenderMode,
    RetextureInfo,
  } from "./types";
  import ModViewerToolbar from "./ModViewerToolbar.svelte";
  import ModViewerStatusBar from "./ModViewerStatusBar.svelte";
  import ModViewerViewportToolbar from "./ModViewerViewportToolbar.svelte";
  import ModViewerCameraPreset from "./ModViewerCameraPreset.svelte";
  import ModViewerTogglesPanel from "./ModViewerTogglesPanel.svelte";
  import { buildAllMeshes } from "./meshFactory";
  import { compact } from "#lib/ts-ext.ts";
  import type { Scene } from "three";

  interface Props {
    mod: ModSaved;
    pack: Pack;
    payload: ViewerPayload;
    onClose: () => void;
    embedded?: boolean;
    isGbPreview?: boolean;
    onInstallGbMod?: () => Promise<void>;
  }

  let {
    mod,
    pack,
    payload,
    onClose,
    embedded = false,
    isGbPreview = false,
    onInstallGbMod,
  }: Props = $props();

  let containerRef: HTMLDivElement | undefined = $state(undefined);

  const scene = $derived(containerRef ? createThreeScene(containerRef) : undefined);

  let loading = $state(true);
  let error = $state<string | null>(null);
  let meshCount = $state(0);
  let builtMeshes = $state<BuiltMesh[]>([]);
  let wireframe = $state(false);
  let turntableActive = $state(false);
  let soloMeshName = $state<string | null>(null);
  let activeCameraPreset = $state<CameraPreset | null>("front");

  let renderMode = $state<RenderMode>("diffuse");
  let lightingPreset = $state<LightingPreset>("studio");

  let toggles = $state<ViewerToggle[]>([]);
  let sidebarTab = $state<"meshes" | "toggles">("meshes");
  let showTogglesPanel = $state(false);
  let potatoMode = $state(true);
  let toggleValues = $state<Record<string, number>>({});

  let installingGb = $state(false);
  let gbInstallSuccess = $state(false);

  let meshVisibility = $state<Record<string, boolean>>({});
  let showBaseModelDropdown = $state(false);

  let retextureInfo = $derived<RetextureInfo>({
    isRetexture: !!payload.isRetexture,
    baseModelSource: payload.baseModelSource || null,
    clues: payload.clues || [],
    candidates: payload.baseModelCandidates || [],
  });

  let preSoloVisibility: Record<string, boolean> = {};
  let totalVertices = $derived(
    builtMeshes.reduce((acc, mesh) => acc + (mesh.vertexCount || 0), 0),
  );
  let totalTriangles = $derived(
    builtMeshes.reduce((acc, mesh) => acc + (mesh.triangleCount || 0), 0),
  );

  const displayName = $derived(pack.name || mod.submission?.name || mod.id);

  function formatCount(value: number): string {
    if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
    if (value >= 1_000) return `${(value / 1_000).toFixed(1)}k`;
    return String(value);
  }

  function initializeViewer(scene: ThreeSceneAPI) {
    try {
      loading = true;
      error = null;

      retextureInfo = {
        isRetexture: !!payload.isRetexture,
        baseModelSource: payload.baseModelSource || null,
        clues: payload.clues || [],
        candidates: payload.baseModelCandidates || [],
      };

      const incomingToggles = payload.toggles || [];
      toggles = incomingToggles;

      const initialValues: Record<string, number> = {};
      for (const toggle of incomingToggles) {
        initialValues[toggle.variable.toLowerCase()] = toggle.currentValue;
      }

      toggleValues = initialValues;

      const meshes = buildAllMeshes(
        payload.meshes ?? [],
        compact(payload.textures ?? {}),
        scene.requestRender,
      );

      builtMeshes = meshes;
      meshCount = meshes.length;

      const visibility: Record<string, boolean> = {};
      for (const mesh of meshes) {
        visibility[mesh.name] = true;
      }

      meshVisibility = visibility;

      scene.setMeshes(meshes);
      scene.applyToggles(initialValues);

      requestAnimationFrame(() => {
        scene.fitCamera();
      });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  function handleSetRenderMode(mode: RenderMode) {
    renderMode = mode;
    scene!.setRenderMode(mode);
  }

  function handleSetLighting(preset: LightingPreset) {
    lightingPreset = preset;
    scene!.setLightingPreset(preset);
  }

  function handleSetToggleValue(variable: string, value: number) {
    const next = {
      ...toggleValues,
      [variable.toLowerCase()]: value,
    };

    toggleValues = next;
    scene!.applyToggles(next);
  }

  function handleCycleToggle(toggle: ViewerToggle) {
    const variable = toggle.variable.toLowerCase();
    const current = toggleValues[variable] ?? toggle.currentValue;
    const index = toggle.values!.indexOf(current);
    const nextIndex = index >= 0 ? (index + 1) % toggle.values!.length : 0;

    const next = {
      ...toggleValues,
      [variable]: toggle.values![nextIndex],
    };

    toggleValues = next;
    scene!.applyToggles(next);
  }

  function toggleWireframe() {
    wireframe = !wireframe;
    scene!.setWireframe(wireframe);
  }

  function toggleMeshVisibility(meshName: string, index: number) {
    const next = !meshVisibility[meshName];

    meshVisibility = {
      ...meshVisibility,
      [meshName]: next,
    };

    scene!.setMeshVisibility(index, next);
  }

  function toggleTurntable() {
    turntableActive = !turntableActive;
    scene!.setTurntable(turntableActive);
  }

  function handleSelectCameraPreset(preset: CameraPreset) {
    activeCameraPreset = preset;
    scene!.setCameraPreset(preset);
  }

  function handleToggleSolo(targetMeshName: string) {
    if (soloMeshName === targetMeshName) {
      const restored = preSoloVisibility;

      builtMeshes.forEach((mesh, index) => {
        scene!.setMeshVisibility(index, restored[mesh.name] ?? true);
      });

      meshVisibility = restored;
      soloMeshName = null;
      return;
    }

    if (!soloMeshName) {
      preSoloVisibility = { ...meshVisibility };
    }

    const nextVisibility: Record<string, boolean> = {};

    builtMeshes.forEach((mesh, index) => {
      const isSolo = mesh.name === targetMeshName;
      nextVisibility[mesh.name] = isSolo;
      scene!.setMeshVisibility(index, isSolo);
    });

    meshVisibility = nextVisibility;
    soloMeshName = targetMeshName;
  }

  function togglePotatoMode() {
    potatoMode = !potatoMode;
  }

  function handleToggleTogglesPanel() {
    showTogglesPanel = !showTogglesPanel;

    if (showTogglesPanel) {
      sidebarTab = "toggles";
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      onClose();
      return;
    }

    if (
      event.key.toLowerCase() === "t" &&
      !(event.target instanceof HTMLInputElement) &&
      !(event.target instanceof HTMLTextAreaElement)
    ) {
      event.preventDefault();
      toggleTurntable();
      return;
    }

    for (const toggle of toggles) {
      if (
        toggle.key &&
        event.key.toLowerCase() === toggle.key.trim().toLowerCase()
      ) {
        event.preventDefault();
        handleCycleToggle(toggle);
        break;
      }
    }
  }

  function handleClickOutside() {
    showBaseModelDropdown = false;
  }

  $effect(() => {
    if (!scene) return
    initializeViewer(scene);

    window.addEventListener("keydown", handleKeyDown);

    return () => window.removeEventListener("keydown", handleKeyDown);
  });
</script>

{#if embedded}
  <div class="relative flex h-full w-full overflow-hidden bg-[#0d1117]">
    <ModViewerToolbar
      isGbPreview={false}
      loadMod={() => {}}
      handleBrowseBaseModel={() => {}}
      {displayName}
      {embedded}
      {retextureInfo}
      {showBaseModelDropdown}
      setShowBaseModelDropdown={(value: boolean) => {
        showBaseModelDropdown = value;
      }}
      {onInstallGbMod}
      {installingGb}
      setInstallingGb={(value: boolean) => {
        installingGb = value;
      }}
      {gbInstallSuccess}
      setGbInstallSuccess={(value: boolean) => {
        gbInstallSuccess = value;
      }}
      {onClose}
    />

    <div class="flex h-full w-full pt-10">
      <div
        bind:this={containerRef}
        class="relative flex-1"
        style="min-height: 0"
      >
        <ModViewerStatusBar
          {retextureInfo}
          onOpenBaseModelDropdown={() => {
            showBaseModelDropdown = true;
          }}
          {loading}
          {error}
          {onClose}
        />

        <ModViewerViewportToolbar
          {meshCount}
          {loading}
          {error}
          {renderMode}
          {handleSetRenderMode}
          {lightingPreset}
          {handleSetLighting}
          {turntableActive}
          {toggleTurntable}
          {wireframe}
          {toggleWireframe}
          rotateModelY={scene!.rotateModelY}
          fitCamera={scene!.fitCamera}
          {potatoMode}
          {togglePotatoMode}
        />

        <ModViewerCameraPreset
          {meshCount}
          {loading}
          {error}
          {activeCameraPreset}
          {handleSelectCameraPreset}
        />

        {#if toggles.length > 0 && !loading && !error}
          <div
            class="absolute bottom-3 right-3 z-20 flex flex-col items-end gap-2"
          >
            {#if showTogglesPanel}
              <ModViewerTogglesPanel
                variant="floating"
                {toggles}
                {toggleValues}
                {handleSetToggleValue}
                {handleCycleToggle}
                onCloseFloating={() => {
                  showTogglesPanel = false;
                }}
              />
            {/if}

            <button
              onclick={handleToggleTogglesPanel}
              class={`flex items-center gap-1.5 rounded-xl border px-3 py-1.5 shadow-2xl backdrop-blur-md text-xs font-semibold transition-all cursor-pointer ${
                showTogglesPanel
                  ? "bg-purple-600 text-white border-purple-400 shadow-purple-900/50"
                  : "bg-[#161b22]/90 text-purple-300 border-purple-500/35 hover:bg-purple-600/30 hover:text-white"
              }`}
            >
              <Sliders class="h-3.5 w-3.5" />
              <span>{toggles.length} Toggles</span>
            </button>
          </div>
        {/if}
      </div>

      {#if builtMeshes.length > 0}
        <div
          class="flex w-72 sm:w-80 flex-col border-l border-white/10 bg-[#0d1117]"
        >
          <div
            class="flex border-b border-white/10 bg-[#161b22]/70 p-1.5 gap-1.5"
          >
            <button
              onclick={() => {
                sidebarTab = "meshes";
              }}
              class={`flex-1 flex items-center justify-center gap-1.5 rounded-md py-1.5 text-xs font-semibold transition-all cursor-pointer ${
                sidebarTab === "meshes"
                  ? "bg-purple-600/30 text-purple-300 border border-purple-500/40 shadow-sm"
                  : "text-white/50 hover:bg-white/5 hover:text-white/80"
              }`}
            >
              <Layers class="h-3.5 w-3.5" />
              <span>Meshes ({builtMeshes.length})</span>
            </button>

            {#if toggles.length > 0}
              <button
                onclick={() => {
                  sidebarTab = "toggles";
                }}
                class={`flex-1 flex items-center justify-center gap-1.5 rounded-md py-1.5 text-xs font-semibold transition-all cursor-pointer ${
                  sidebarTab === "toggles"
                    ? "bg-purple-600/30 text-purple-300 border border-purple-500/40 shadow-sm"
                    : "text-white/50 hover:bg-white/5 hover:text-white/80"
                }`}
              >
                <Sliders class="h-3.5 w-3.5" />
                <span>Toggles ({toggles.length})</span>
              </button>
            {/if}
          </div>

          {#if sidebarTab === "meshes"}
            <div></div>
          {:else}
            <ModViewerTogglesPanel
              variant="sidebar"
              {toggles}
              {toggleValues}
              {handleSetToggleValue}
              {handleCycleToggle}
            />
          {/if}
        </div>
      {/if}
    </div>
  </div>
{:else}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-2 sm:p-4 animate-in fade-in duration-200"
  >
    <div
      class="relative flex h-full max-h-[92vh] w-full max-w-7xl flex-col overflow-hidden rounded-2xl border border-white/10 shadow-2xl"
    >
      <div class="relative flex h-full w-full overflow-hidden bg-[#0d1117]">
        <ModViewerToolbar
          isGbPreview={false}
          loadMod={() => {}}
          handleBrowseBaseModel={() => {}}
          {displayName}
          {embedded}
          {retextureInfo}
          {showBaseModelDropdown}
          setShowBaseModelDropdown={(value: boolean) => {
            showBaseModelDropdown = value;
          }}
          {onInstallGbMod}
          {installingGb}
          setInstallingGb={(value: boolean) => {
            installingGb = value;
          }}
          {gbInstallSuccess}
          setGbInstallSuccess={(value: boolean) => {
            gbInstallSuccess = value;
          }}
          {onClose}
        />

        <div class="flex h-full w-full pt-10">
          <div
            bind:this={containerRef}
            class="relative flex-1"
            style="min-height: 0"
          >
            <ModViewerStatusBar
              {retextureInfo}
              onOpenBaseModelDropdown={() => {
                showBaseModelDropdown = true;
              }}
              {loading}
              {error}
              {onClose}
            />

            <ModViewerViewportToolbar
              {meshCount}
              {loading}
              {error}
              {renderMode}
              {handleSetRenderMode}
              {lightingPreset}
              {handleSetLighting}
              {turntableActive}
              {toggleTurntable}
              {wireframe}
              {toggleWireframe}
              rotateModelY={scene!.rotateModelY}
              fitCamera={scene!.fitCamera}
              {potatoMode}
              {togglePotatoMode}
            />

            <ModViewerCameraPreset
              {meshCount}
              {loading}
              {error}
              {activeCameraPreset}
              {handleSelectCameraPreset}
            />

            {#if toggles.length > 0 && !loading && !error}
              <div
                class="absolute bottom-3 right-3 z-20 flex flex-col items-end gap-2"
              >
                {#if showTogglesPanel}
                  <ModViewerTogglesPanel
                    variant="floating"
                    {toggles}
                    {toggleValues}
                    {handleSetToggleValue}
                    {handleCycleToggle}
                    onCloseFloating={() => {
                      showTogglesPanel = false;
                    }}
                  />
                {/if}

                <button
                  onclick={handleToggleTogglesPanel}
                  class={`flex items-center gap-1.5 rounded-xl border px-3 py-1.5 shadow-2xl backdrop-blur-md text-xs font-semibold transition-all cursor-pointer ${
                    showTogglesPanel
                      ? "bg-purple-600 text-white border-purple-400 shadow-purple-900/50"
                      : "bg-[#161b22]/90 text-purple-300 border-purple-500/35 hover:bg-purple-600/30 hover:text-white"
                  }`}
                >
                  <Sliders class="h-3.5 w-3.5" />
                  <span>{toggles.length} Toggles</span>
                </button>
              </div>
            {/if}
          </div>

          {#if builtMeshes.length > 0}
            <div
              class="flex w-72 sm:w-80 flex-col border-l border-white/10 bg-[#0d1117]"
            >
              <div
                class="flex border-b border-white/10 bg-[#161b22]/70 p-1.5 gap-1.5"
              >
                <button
                  onclick={() => {
                    sidebarTab = "meshes";
                  }}
                  class={`flex-1 flex items-center justify-center gap-1.5 rounded-md py-1.5 text-xs font-semibold transition-all cursor-pointer ${
                    sidebarTab === "meshes"
                      ? "bg-purple-600/30 text-purple-300 border border-purple-500/40 shadow-sm"
                      : "text-white/50 hover:bg-white/5 hover:text-white/80"
                  }`}
                >
                  <Layers class="h-3.5 w-3.5" />
                  <span>Meshes ({builtMeshes.length})</span>
                </button>

                {#if toggles.length > 0}
                  <button
                    onclick={() => {
                      sidebarTab = "toggles";
                    }}
                    class={`flex-1 flex items-center justify-center gap-1.5 rounded-md py-1.5 text-xs font-semibold transition-all cursor-pointer ${
                      sidebarTab === "toggles"
                        ? "bg-purple-600/30 text-purple-300 border border-purple-500/40 shadow-sm"
                        : "text-white/50 hover:bg-white/5 hover:text-white/80"
                    }`}
                  >
                    <Sliders class="h-3.5 w-3.5" />
                    <span>Toggles ({toggles.length})</span>
                  </button>
                {/if}
              </div>

              {#if sidebarTab === "meshes"}
                <!-- <ModViewerComponentsPanel
                  {builtMeshes}
                  {meshVisibility}
                  {toggleMeshVisibility}
                  {soloMeshName}
                  {handleToggleSolo}
                  {totalTriangles}
                  {totalVertices}
                  {formatCount}
                /> -->
              {:else}
                <ModViewerTogglesPanel
                  variant="sidebar"
                  {toggles}
                  {toggleValues}
                  {handleSetToggleValue}
                  {handleCycleToggle}
                />
              {/if}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
