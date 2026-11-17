
<script lang="ts">
  import * as THREE from "three";
  import { OrbitControls } from "three/addons/controls/OrbitControls.js";
  import { GLTFLoader } from "three/addons/loaders/GLTFLoader.js";

  let { preview }: { preview: string | null } = $props();

  let canvas = $state<HTMLCanvasElement>();
  let ready = $state(false);
  let loading = $state(false);
  let error = $state<string | null>(null);

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(35, 1, 0.01, 1000);

  let renderer: THREE.WebGLRenderer | undefined;
  let controls: OrbitControls | undefined;
  let model: THREE.Group | undefined;

  const loader = new GLTFLoader();

  const fallbackMaterial = new THREE.MeshStandardMaterial({
    color: 0xb8a99a,
    roughness: 0.85,
    metalness: 0,
    side: THREE.DoubleSide,
  });

  function render() {
    renderer?.render(scene, camera);
  }

  function resize() {
    if (!renderer || !canvas) return;

    const width = canvas.clientWidth;
    const height = canvas.clientHeight;

    if (!width || !height) return;

    renderer.setSize(width, height, false);

    camera.aspect = width / height;
    camera.updateProjectionMatrix();

    render();
  }

  function frame(root: THREE.Object3D) {
    const box = new THREE.Box3().setFromObject(root);

    if (box.isEmpty() || !controls) {
      console.warn("[Preview] Cannot frame model: bounding box is empty or controls unavailable");
      return;
    }

    const center = box.getCenter(new THREE.Vector3());
    const size = box.getSize(new THREE.Vector3());
    const radius = size.length() / 2 || 1;

    const distance =
      (radius / Math.sin(THREE.MathUtils.degToRad(camera.fov) / 2)) * 1.05;

    console.groupCollapsed("[Preview] Camera framing");
    console.log("Bounds:", {
      min: box.min.toArray(),
      max: box.max.toArray(),
    });
    console.log("Size:", size.toArray());
    console.log("Center:", center.toArray());
    console.log("Camera distance:", distance);
    console.groupEnd();

    camera.position.set(center.x, center.y, center.z - distance);

    camera.near = distance / 100;
    camera.far = distance * 100;
    camera.updateProjectionMatrix();

    controls.target.copy(center);
    controls.minDistance = distance * 0.2;
    controls.maxDistance = distance * 5;
    controls.update();
  }

  function getTextureInfo(texture: THREE.Texture | null | undefined) {
    if (!texture) return null;

    const image = texture.image as
      | { width?: number; height?: number }
      | undefined;

    return {
      uuid: texture.uuid,
      name: texture.name,
      type: texture.type,
      format: texture.format,
      colorSpace: texture.colorSpace,
      mapping: texture.mapping,
      wrapS: texture.wrapS,
      wrapT: texture.wrapT,
      flipY: texture.flipY,
      generateMipmaps: texture.generateMipmaps,
      minFilter: texture.minFilter,
      magFilter: texture.magFilter,
      dimensions: image
        ? `${image.width ?? "?"}x${image.height ?? "?"}`
        : "Unknown",
      image,
      isDataTexture: texture instanceof THREE.DataTexture,
      isCompressedTexture: texture instanceof THREE.CompressedTexture,
    };
  }

  function getMaterialInfo(material: THREE.Material) {
    const info: Record<string, unknown> = {
      name: material.name,
      type: material.type,
      uuid: material.uuid,
      visible: material.visible,
      opacity: material.opacity,
      transparent: material.transparent,
      side: material.side,
      depthWrite: material.depthWrite,
      depthTest: material.depthTest,
      blending: material.blending,
    };

    if (material instanceof THREE.MeshStandardMaterial) {
      Object.assign(info, {
        color: `#${material.color.getHexString()}`,
        roughness: material.roughness,
        metalness: material.metalness,
        emissive: `#${material.emissive.getHexString()}`,
        emissiveIntensity: material.emissiveIntensity,
        map: getTextureInfo(material.map),
        normalMap: getTextureInfo(material.normalMap),
        roughnessMap: getTextureInfo(material.roughnessMap),
        metalnessMap: getTextureInfo(material.metalnessMap),
        emissiveMap: getTextureInfo(material.emissiveMap),
        aoMap: getTextureInfo(material.aoMap),
        alphaMap: getTextureInfo(material.alphaMap),
      });
    } else if (material instanceof THREE.MeshBasicMaterial) {
      Object.assign(info, {
        color: `#${material.color.getHexString()}`,
        map: getTextureInfo(material.map),
        alphaMap: getTextureInfo(material.alphaMap),
      });
    } else if (material instanceof THREE.MeshPhongMaterial) {
      Object.assign(info, {
        color: `#${material.color.getHexString()}`,
        specular: `#${material.specular.getHexString()}`,
        shininess: material.shininess,
        map: getTextureInfo(material.map),
        normalMap: getTextureInfo(material.normalMap),
      });
    }

    const textures: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(material)) {
      if (value instanceof THREE.Texture) {
        textures[key] = getTextureInfo(value);
      }
    }

    info.textures = textures;

    return info;
  }

  function prepareModel(root: THREE.Object3D) {
    const stats = {
      meshes: 0,
      vertices: 0,
      indices: 0,
      missingUV: 0,
      missingNormals: 0,
      missingColors: 0,
      missingTextures: 0,
      fallbackMaterials: 0,
    };

    console.group("[Preview] Mesh diagnostics");

    root.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) return;

      stats.meshes++;

      const geometry = object.geometry;
      const position = geometry.getAttribute("position");
      const uv = geometry.getAttribute("uv");
      const normal = geometry.getAttribute("normal");
      const tangent = geometry.getAttribute("tangent");
      const color = geometry.getAttribute("color");

      stats.vertices += position?.count ?? 0;
      stats.indices += geometry.index?.count ?? 0;

      if (!uv) stats.missingUV++;
      if (!normal) stats.missingNormals++;
      if (!color) stats.missingColors++;

      if (!normal) {
        geometry.computeVertexNormals();
      }

      const materials = Array.isArray(object.material)
        ? object.material
        : [object.material];

      const materialInfo = materials.map(getMaterialInfo);

      const hasTexture = materials.some((material) => {
        if (material instanceof THREE.MeshStandardMaterial) {
          return !!material.map;
        }

        if (material instanceof THREE.MeshBasicMaterial) {
          return !!material.map;
        }

        return false;
      });

      if (!hasTexture) stats.missingTextures++;

      const hasUsableMaterial = materials.some((material) => {
        if (!material) return false;

        if (material instanceof THREE.MeshStandardMaterial) {
          return (
            material.map !== null ||
            material.color.getHex() !== 0xffffff
          );
        }

        if (material instanceof THREE.MeshBasicMaterial) {
          return (
            material.map !== null ||
            material.color.getHex() !== 0xffffff
          );
        }

        return true;
      });

      if (!hasUsableMaterial) {
        object.material = fallbackMaterial;
        stats.fallbackMaterials++;
      }

      const bounds = new THREE.Box3().setFromBufferAttribute(
        position as THREE.BufferAttribute
      );

      console.groupCollapsed(
        `[Preview] Mesh ${stats.meshes}: ${object.name || "(unnamed)"}`
      );

      console.log("Object:", object);
      console.log("Geometry:", geometry);
      console.log("Material:", object.material);

      console.log("Geometry summary:", {
        name: geometry.name,
        vertices: position?.count,
        indices: geometry.index?.count,
        triangles: geometry.index
          ? geometry.index.count / 3
          : (position?.count ?? 0) / 3,
        groups: geometry.groups,
        drawRange: geometry.drawRange,
        boundingBox: {
          min: bounds.min.toArray(),
          max: bounds.max.toArray(),
        },
      });

      console.log("Attributes:", {
        position: position
          ? {
              count: position.count,
              itemSize: position.itemSize,
              normalized: position.normalized,
              arrayType: position.array.constructor.name,
              firstValues: Array.from(position.array.slice(0, 12)),
            }
          : null,
        uv: uv
          ? {
              count: uv.count,
              itemSize: uv.itemSize,
              normalized: uv.normalized,
              arrayType: uv.array.constructor.name,
              firstValues: Array.from(uv.array.slice(0, 12)),
            }
          : null,
        normal: normal
          ? {
              count: normal.count,
              itemSize: normal.itemSize,
              firstValues: Array.from(normal.array.slice(0, 12)),
            }
          : null,
        tangent: tangent
          ? {
              count: tangent.count,
              itemSize: tangent.itemSize,
              firstValues: Array.from(tangent.array.slice(0, 12)),
            }
          : null,
        color: color
          ? {
              count: color.count,
              itemSize: color.itemSize,
              normalized: color.normalized,
              firstValues: Array.from(color.array.slice(0, 12)),
            }
          : null,
        attributes: Object.keys(geometry.attributes),
      });

      console.log("Materials:", materialInfo);

      console.log("Material diagnostics:", {
        hasTexture,
        materialCount: materials.length,
        fallbackApplied: !hasUsableMaterial,
      });

      console.groupEnd();
    });

    console.groupEnd();

    console.group("[Preview] Model summary");
    console.table(stats);
    console.log("Scene:", root);
    console.log("Scene children:", root.children);
    console.groupEnd();
  }

  function disposeModel(root: THREE.Object3D) {
    const geometries = new Set<THREE.BufferGeometry>();
    const materials = new Set<THREE.Material>();
    const textures = new Set<THREE.Texture>();

    root.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) return;

      geometries.add(object.geometry);

      const meshMaterials = Array.isArray(object.material)
        ? object.material
        : [object.material];

      for (const material of meshMaterials) {
        materials.add(material);

        for (const value of Object.values(material)) {
          if (value instanceof THREE.Texture) {
            textures.add(value);
          }
        }
      }
    });

    console.log("[Preview] Disposing model resources:", {
      geometries: geometries.size,
      materials: materials.size,
      textures: textures.size,
    });

    geometries.forEach((geometry) => geometry.dispose());
    materials.forEach((material) => material.dispose());
    textures.forEach((texture) => texture.dispose());
  }

  function clearModel() {
    if (!model) return;

    console.log("[Preview] Clearing previous model");

    scene.remove(model);
    disposeModel(model);
    model = undefined;
  }

  async function loadPreview(path: string) {
    if (!renderer || !controls) return;

    loading = true;
    error = null;

    clearModel();

    const startedAt = performance.now();

    console.group("[Preview] Loading GLB");
    console.log("Path:", path);

    try {
      const gltf = await loader.loadAsync(path);

      console.log("GLB loaded:", {
        durationMs: Math.round(performance.now() - startedAt),
        animations: gltf.animations.length,
        scenes: gltf.scenes.length,
        activeScene: gltf.scene,
        asset: gltf.asset,
        parser: gltf.parser,
      });

      if (!renderer || !controls) {
        disposeModel(gltf.scene);
        return;
      }

      model = gltf.scene;

      prepareModel(model);

      scene.add(model);

      frame(model);
      render();

      console.log("[Preview] Model rendered successfully");
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
      console.error("[Preview] Failed to load preview:", e);
    } finally {
      loading = false;
      console.groupEnd();
    }
  }

  $effect(() => {
    if (!canvas) return;

    const r = new THREE.WebGLRenderer({
      canvas,
      antialias: true,
      alpha: true,
    });

    r.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    r.outputColorSpace = THREE.SRGBColorSpace;
    r.toneMapping = THREE.ACESFilmicToneMapping;
    r.toneMappingExposure = 1;
    renderer = r;

    const c = new OrbitControls(camera, canvas);
    c.enableDamping = true;
    c.addEventListener("change", render);
    controls = c;

    scene.add(new THREE.HemisphereLight(0xffffff, 0x777788, 2));

    const key = new THREE.DirectionalLight(0xffffff, 2.5);
    key.position.set(5, 8, 5);
    scene.add(key);

    const fill = new THREE.DirectionalLight(0xffffff, 1.5);
    fill.position.set(-5, 3, -5);
    scene.add(fill);

    const observer = new ResizeObserver(resize);
    observer.observe(canvas);

    console.log("[Preview] Renderer initialized:", {
      threeVersion: THREE.REVISION,
      pixelRatio: r.getPixelRatio(),
      outputColorSpace: r.outputColorSpace,
      toneMapping: r.toneMapping,
      maxTextureSize: r.capabilities.maxTextureSize,
      maxAttributes: r.capabilities.maxAttributes,
      webgl2: r.capabilities.isWebGL2,
    });

    resize();
    ready = true;

    return () => {
      ready = false;

      observer.disconnect();
      c.dispose();

      clearModel();

      fallbackMaterial.dispose();

      scene.clear();

      r.dispose();

      renderer = undefined;
      controls = undefined;
    };
  });

  $effect(() => {
    const path = preview;

    if (!ready || !path) {
      clearModel();
      error = null;
      render();
      return;
    }

    void loadPreview(path);
  });
</script>

<div class="relative h-full w-full">
  <canvas bind:this={canvas} class="block h-full w-full"></canvas>

  {#if loading}
    <div
      class="absolute inset-0 flex items-center justify-center text-sm text-muted-foreground"
    >
      Loading preview...
    </div>
  {:else if error}
    <div
      class="absolute inset-0 flex items-center justify-center p-4 text-center text-sm text-destructive"
    >
      {error}
    </div>
  {:else if !preview}
    <div
      class="absolute inset-0 flex items-center justify-center text-sm text-muted-foreground"
    >
      Nothing to preview
    </div>
  {/if}
</div>
