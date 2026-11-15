<script lang="ts">
  import type { MeshData } from "#lib/bindings/changeme/models.ts";
  import * as THREE from "three";
  import { OrbitControls } from "./OrbitalControls";

  type ModelProps = {
    mesh: MeshData | null;
  };

  let { mesh }: ModelProps = $props();

  function renderMesh(canvas: HTMLCanvasElement, mesh: MeshData) {
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x202020);

    const camera = new THREE.PerspectiveCamera(
      50,
      canvas.clientWidth / canvas.clientHeight,
      0.1,
      1000,
    );

    camera.position.set(0, 1, 4);

    const renderer = new THREE.WebGLRenderer({
      canvas,
      antialias: true,
    });

    renderer.setSize(canvas.clientWidth, canvas.clientHeight);
    renderer.setPixelRatio(window.devicePixelRatio);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;

    scene.add(new THREE.AmbientLight(0xffffff, 1.5));

    const light = new THREE.DirectionalLight(0xffffff, 2);
    light.position.set(5, 5, 5);
    scene.add(light);

    const geometry = createGeometry(mesh);

    const material = new THREE.MeshBasicMaterial({
      color: 0xffffff,
      wireframe: true,
      side: THREE.DoubleSide,
    });

    const object = new THREE.Mesh(geometry, material);
    scene.add(object);

    const animate = () => {
      renderer.render(scene, camera);
    };

    renderer.setAnimationLoop(animate);

    return () => {
      renderer.setAnimationLoop(null);
      geometry.dispose();
      material.dispose();
      renderer.dispose();
    };
  }

  function createGeometry(mesh: MeshData): THREE.BufferGeometry {
    const geometry = new THREE.BufferGeometry();

    geometry.setAttribute(
      "position",
      new THREE.Float32BufferAttribute(mesh.positions!, 3),
    );

    geometry.setAttribute(
      "normal",
      new THREE.Float32BufferAttribute(mesh.normals!, 3),
    );

    geometry.setAttribute(
      "tangent",
      new THREE.Float32BufferAttribute(mesh.tangents!, 4),
    );

    geometry.setAttribute("uv", new THREE.Float32BufferAttribute(mesh.uvs!, 2));

    geometry.setAttribute(
      "color",
      new THREE.Float32BufferAttribute(mesh.colors!, 4),
    );

    geometry.setIndex(mesh.indices);

    geometry.computeBoundingBox();
    geometry.computeBoundingSphere();

    return geometry;
  }

  let canvas: HTMLCanvasElement;
  let cleanup: (() => void) | undefined;

  $effect(() => {
    if (!canvas || !mesh) return;

    cleanup = renderMesh(canvas, mesh);

    return () => cleanup?.();
  });
</script>

<div class="h-full w-full">
  <canvas bind:this={canvas} class="h-full w-full"></canvas>
</div>
