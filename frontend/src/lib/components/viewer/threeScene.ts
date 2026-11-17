
import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import type { BuiltMesh } from './types';

export type RenderMode = 'pbr' | 'diffuse' | 'clay' | 'normal';
export type LightingPreset = 'studio' | 'anime' | 'dramatic';
export type CameraPreset = 'front' | 'back' | 'left' | 'right' | 'face' | 'reset';

export interface ThreeSceneAPI {
	setMeshes: (meshes: BuiltMesh[]) => void;
	clear: () => void;
	requestRender: () => void;
	resize: () => void;
	fitCamera: () => void;
	setCameraPreset: (preset: CameraPreset) => void;
	setTurntable: (enabled: boolean) => void;
	rotateModelY: (angle?: number) => void;
	setRenderMode: (mode: RenderMode) => void;
	setWireframe: (enabled: boolean) => void;
	setLightingPreset: (preset: LightingPreset) => void;
	setMeshVisibility: (index: number, visible: boolean) => void;
	applyToggles: (toggleValues: Record<string, number>) => void;
	captureSnapshot: () => string | null;
	destroy: () => void;
}

interface ModeMaterials {
	diffuse?: THREE.MeshStandardMaterial;
	clay?: THREE.MeshStandardMaterial;
	normal?: THREE.MeshNormalMaterial;
}

export function evaluateCondition(
	expr: string | null | undefined,
	values: Record<string, number>
): boolean {
	if (!expr || !expr.trim()) return true;

	const normalizedValues: Record<string, number> = {};

	for (const [key, value] of Object.entries(values)) {
		const lowerKey = key.toLowerCase().trim();
		const stripped = lowerKey.startsWith('$') ? lowerKey.slice(1) : lowerKey;

		normalizedValues[lowerKey] = value;
		normalizedValues[stripped] = value;
		normalizedValues[`$${stripped}`] = value;
	}

	const resolveToken = (token: string): number => {
		const raw = token.trim();
		if (!raw) return 0;

		const lower = raw.toLowerCase();

		if (lower in normalizedValues) return normalizedValues[lower];

		const stripped = lower.startsWith('$') ? lower.slice(1) : lower;
		if (stripped in normalizedValues) return normalizedValues[stripped];

		const num = Number(raw);
		return Number.isFinite(num) ? num : 0;
	};

	const evaluateClause = (clause: string): boolean => {
		const trimmed = clause.trim().replace(/^\(|\)$/g, '').trim();
		if (!trimmed) return true;

		for (const op of ['>=', '<=', '!=', '==', '>', '<'] as const) {
			if (!trimmed.includes(op)) continue;

			const [left, right] = trimmed.split(op).map((s) => s.trim());
			const leftValue = resolveToken(left);
			const rightValue = resolveToken(right);

			switch (op) {
				case '==':
					return leftValue === rightValue;
				case '!=':
					return leftValue !== rightValue;
				case '>=':
					return leftValue >= rightValue;
				case '<=':
					return leftValue <= rightValue;
				case '>':
					return leftValue > rightValue;
				case '<':
					return leftValue < rightValue;
			}
		}

		if (trimmed.startsWith('!')) {
			return resolveToken(trimmed.slice(1)) === 0;
		}

		return resolveToken(trimmed) !== 0;
	};

	const orBranches = expr
		.split('||')
		.map((branch) => branch.trim())
		.filter(Boolean);

	return orBranches.some((branch) => {
		const andClauses = branch
			.split('&&')
			.map((clause) => clause.trim())
			.filter(Boolean);

		return andClauses.every(evaluateClause);
	});
}

function applyMaterialMode(
	mesh: THREE.Mesh,
	mode: RenderMode,
	wireframe: boolean
) {
	const original = mesh.userData.originalMaterial as
		| THREE.MeshStandardMaterial
		| undefined;

	if (!original) return;

	const cache = (mesh.userData.modeMaterials ??= {}) as ModeMaterials;

	switch (mode) {
		case 'pbr':
			original.wireframe = wireframe;
			mesh.material = original;
			break;

		case 'diffuse':
			cache.diffuse ??= new THREE.MeshStandardMaterial({
				color: 0xffffff,
				map: mesh.userData.diffuseTexture || null,
				side: THREE.DoubleSide,
				roughness: 0.85,
				metalness: 0
			});

			cache.diffuse.wireframe = wireframe;
			mesh.material = cache.diffuse;
			break;

		case 'clay':
			cache.clay ??= new THREE.MeshStandardMaterial({
				color: 0xc4c4c4,
				side: THREE.DoubleSide,
				roughness: 0.55,
				metalness: 0
			});

			cache.clay.wireframe = wireframe;
			mesh.material = cache.clay;
			break;

		case 'normal':
			cache.normal ??= new THREE.MeshNormalMaterial({
				side: THREE.DoubleSide
			});

			cache.normal.wireframe = wireframe;
			mesh.material = cache.normal;
			break;
	}
}

function disposeMaterial(material: THREE.Material) {
	if ('map' in material && material.map instanceof THREE.Texture) {
		material.map.dispose();
	}

	material.dispose();
}

function disposeThreeMesh(object: THREE.Object3D) {
	if (!(object instanceof THREE.Mesh)) return;

	object.geometry.dispose();

	const modeMaterials = object.userData.modeMaterials as
		| ModeMaterials
		| undefined;

	modeMaterials?.diffuse?.dispose();
	modeMaterials?.clay?.dispose();
	modeMaterials?.normal?.dispose();

	const original = object.userData.originalMaterial as
		| THREE.Material
		| undefined;

	const current = object.material;

	if (Array.isArray(current)) {
		for (const material of current) {
			if (material !== original) disposeMaterial(material);
		}
	} else if (current !== original) {
		disposeMaterial(current);
	}

	if (original) {
		disposeMaterial(original);
	}

	for (const key of [
		'diffuseTexture',
		'normalTexture',
		'lightTexture',
		'materialTexture'
	]) {
		const texture = object.userData[key] as THREE.Texture | undefined;
		texture?.dispose();
	}
}

function clearGroup(group: THREE.Group) {
	while (group.children.length > 0) {
		const child = group.children[0];
		group.remove(child);
		disposeThreeMesh(child);
	}
}

function getMeshBounds(group: THREE.Group): THREE.Box3 {
	group.updateMatrixWorld(true);

	const box = new THREE.Box3();

	group.traverse((child) => {
		if (!(child instanceof THREE.Mesh)) return;

		child.geometry.computeBoundingBox();

		const childBox = child.geometry.boundingBox;
		if (!childBox) return;

		const worldBox = childBox.clone().applyMatrix4(child.matrixWorld);
		box.union(worldBox);
	});

	return box;
}

export function createThreeScene(container: HTMLElement): ThreeSceneAPI {
	let destroyed = false;
	let animFrame = 0;
	let needsRender = true;

	let renderMode: RenderMode = 'diffuse';
	let wireframe = false;
	let turntable = false;

	const renderer = new THREE.WebGLRenderer({
		antialias: true,
		alpha: true,
		preserveDrawingBuffer: true
	});

	renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
	renderer.setSize(
		Math.max(1, container.clientWidth),
		Math.max(1, container.clientHeight)
	);
	renderer.outputColorSpace = THREE.SRGBColorSpace;
	renderer.toneMapping = THREE.ACESFilmicToneMapping;
	renderer.toneMappingExposure = 1;
	renderer.setClearColor(0x1a1a2e, 1);

	container.appendChild(renderer.domElement);

	const scene = new THREE.Scene();

	const meshGroup = new THREE.Group();
	scene.add(meshGroup);

	const camera = new THREE.PerspectiveCamera(
		50,
		Math.max(1, container.clientWidth) /
			Math.max(1, container.clientHeight),
		0.01,
		1000
	);

	camera.position.set(0, 1.5, 3);

	const controls = new OrbitControls(camera, renderer.domElement);
	controls.enableDamping = true;
	controls.dampingFactor = 0.1;
	controls.target.set(0, 1, 0);
	controls.update();

	const ambientLight = new THREE.AmbientLight(0xffffff, 0.85);
	scene.add(ambientLight);

	const dirLight = new THREE.DirectionalLight(0xffffff, 0.9);
	dirLight.position.set(3, 8, 6);
	dirLight.castShadow = false;
	scene.add(dirLight);

	const fillLight = new THREE.DirectionalLight(0xeeeeff, 0.45);
	fillLight.position.set(-4, 4, -4);
	scene.add(fillLight);

	const hemiLight = new THREE.HemisphereLight(0xffffff, 0x333344, 0.45);
	scene.add(hemiLight);

	const grid = new THREE.GridHelper(10, 20, 0x444444, 0x333333);
	scene.add(grid);

	const onControlsChange = () => {
		needsRender = true;
	};

	controls.addEventListener('change', onControlsChange);

	function requestRender() {
		needsRender = true;
	}

	function resize() {
		if (destroyed) return;

		const width = container.clientWidth;
		const height = container.clientHeight;

		if (width <= 0 || height <= 0) return;

		camera.aspect = width / height;
		camera.updateProjectionMatrix();

		renderer.setSize(width, height);
		requestRender();
	}

	function animate() {
		if (destroyed) return;

		animFrame = requestAnimationFrame(animate);

		if (turntable) {
			controls.autoRotate = true;
			controls.autoRotateSpeed = 2;
			needsRender = true;
		} else {
			controls.autoRotate = false;
		}

		controls.update();

		if (needsRender) {
			renderer.render(scene, camera);
			needsRender = false;
		}
	}

	const resizeObserver = new ResizeObserver(resize);
	resizeObserver.observe(container);

	animate();

	function setMeshes(builtMeshes: BuiltMesh[]) {
		if (destroyed) return;

		meshGroup.rotation.set(0, 0, 0);
		clearGroup(meshGroup);

		for (const built of builtMeshes) {
			applyMaterialMode(built.mesh, renderMode, wireframe);
			meshGroup.add(built.mesh);
		}

		requestRender();
	}

	function clear() {
		setMeshes([]);
	}

	function fitCamera() {
		if (destroyed || meshGroup.children.length === 0) return;

		const box = getMeshBounds(meshGroup);
		if (box.isEmpty()) return;

		const center = box.getCenter(new THREE.Vector3());
		const size = box.getSize(new THREE.Vector3());
		const maxDim = Math.max(size.x, size.y, size.z);
		const fov = camera.fov * (Math.PI / 180);
		const distance = (maxDim / (2 * Math.tan(fov / 2))) * 1.35;

		grid.position.set(center.x, box.min.y, center.z);

		const gridScale = Math.max(1, maxDim * 0.4);
		grid.scale.set(gridScale, gridScale, gridScale);

		camera.position.set(
			center.x,
			center.y + size.y * 0.1,
			center.z + distance
		);

		controls.target.copy(center);

		camera.near = distance * 0.01;
		camera.far = distance * 10;
		camera.updateProjectionMatrix();

		controls.update();
		requestRender();
	}

	function setCameraPreset(preset: CameraPreset) {
		if (preset === 'reset') {
			fitCamera();
			return;
		}

		if (destroyed || meshGroup.children.length === 0) return;

		const box = getMeshBounds(meshGroup);
		if (box.isEmpty()) return;

		const center = box.getCenter(new THREE.Vector3());
		const size = box.getSize(new THREE.Vector3());
		const maxDim = Math.max(size.x, size.y, size.z);
		const fov = camera.fov * (Math.PI / 180);
		const distance = (maxDim / (2 * Math.tan(fov / 2))) * 1.35;
		const target = center.clone();

		switch (preset) {
			case 'front':
				camera.position.set(
					center.x,
					center.y + size.y * 0.05,
					center.z + distance
				);
				break;

			case 'back':
				camera.position.set(
					center.x,
					center.y + size.y * 0.05,
					center.z - distance
				);
				break;

			case 'left':
				camera.position.set(
					center.x - distance,
					center.y + size.y * 0.05,
					center.z
				);
				break;

			case 'right':
				camera.position.set(
					center.x + distance,
					center.y + size.y * 0.05,
					center.z
				);
				break;

			case 'face': {
				const faceY = box.max.y - size.y * 0.15;
				target.set(center.x, faceY, center.z);

				const faceDistance = Math.max(0.4, distance * 0.32);
				camera.position.set(
					center.x,
					faceY,
					center.z + faceDistance
				);
				break;
			}
		}

		controls.target.copy(target);
		controls.update();
		requestRender();
	}

	function setTurntable(enabled: boolean) {
		turntable = enabled;
		controls.autoRotate = enabled;
		controls.autoRotateSpeed = 2;
		requestRender();
	}

	function rotateModelY(angle = Math.PI / 2) {
		if (destroyed) return;

		meshGroup.rotation.y += angle;
		requestRender();
	}

	function setRenderMode(mode: RenderMode) {
		renderMode = mode;

		for (const child of meshGroup.children) {
			if (child instanceof THREE.Mesh) {
				applyMaterialMode(child, mode, wireframe);
			}
		}

		requestRender();
	}

	function setWireframe(enabled: boolean) {
		wireframe = enabled;

		for (const child of meshGroup.children) {
			if (child instanceof THREE.Mesh) {
				applyMaterialMode(child, renderMode, enabled);
			}
		}

		requestRender();
	}

	function setLightingPreset(preset: LightingPreset) {
		if (destroyed) return;

		switch (preset) {
			case 'studio':
				ambientLight.color.setHex(0xffffff);
				ambientLight.intensity = 0.6;

				dirLight.color.setHex(0xffffff);
				dirLight.intensity = 1.2;
				dirLight.position.set(5, 10, 7);

				fillLight.color.setHex(0x8888ff);
				fillLight.intensity = 0.4;

				hemiLight.color.setHex(0xb1e1ff);
				hemiLight.groundColor.setHex(0x444444);
				hemiLight.intensity = 0.5;
				break;

			case 'anime':
				ambientLight.color.setHex(0xffffff);
				ambientLight.intensity = 1.1;

				dirLight.color.setHex(0xffffff);
				dirLight.intensity = 0.7;
				dirLight.position.set(2, 6, 6);

				fillLight.color.setHex(0xffffff);
				fillLight.intensity = 0.5;

				hemiLight.color.setHex(0xffffff);
				hemiLight.groundColor.setHex(0x888888);
				hemiLight.intensity = 0.7;
				break;

			case 'dramatic':
				ambientLight.color.setHex(0x222233);
				ambientLight.intensity = 0.3;

				dirLight.color.setHex(0xffffff);
				dirLight.intensity = 2;
				dirLight.position.set(6, 8, 4);

				fillLight.color.setHex(0x3355aa);
				fillLight.intensity = 0.8;

				hemiLight.color.setHex(0x223344);
				hemiLight.groundColor.setHex(0x111122);
				hemiLight.intensity = 0.2;
				break;
		}

		requestRender();
	}

	function setMeshVisibility(index: number, visible: boolean) {
		if (destroyed || index < 0 || index >= meshGroup.children.length) return;

		meshGroup.children[index].visible = visible;
		requestRender();
	}

	function applyToggles(toggleValues: Record<string, number>) {
		if (destroyed) return;

		for (const child of meshGroup.children) {
			if (!(child instanceof THREE.Mesh)) continue;

			const condition = child.userData.condition as string | undefined;

			if (condition) {
				child.visible = evaluateCondition(condition, toggleValues);
			}
		}

		requestRender();
	}

	function captureSnapshot(): string | null {
		if (destroyed) return null;

		const wasGridVisible = grid.visible;
		grid.visible = false;

		renderer.render(scene, camera);

		let dataUrl: string | null = null;

		try {
			dataUrl = renderer.domElement.toDataURL('image/png');
		} catch {
			dataUrl = null;
		} finally {
			grid.visible = wasGridVisible;
			requestRender();
		}

		return dataUrl;
	}

	function destroy() {
		if (destroyed) return;

		destroyed = true;

		resizeObserver.disconnect();
		cancelAnimationFrame(animFrame);

		controls.removeEventListener('change', onControlsChange);
		controls.dispose();

		clearGroup(meshGroup);

		grid.geometry.dispose();

		if (Array.isArray(grid.material)) {
			grid.material.forEach((material) => material.dispose());
		} else {
			grid.material.dispose();
		}

		ambientLight.dispose();
		dirLight.dispose();
		fillLight.dispose();
		hemiLight.dispose();

		renderer.dispose();

		if (container.contains(renderer.domElement)) {
			container.removeChild(renderer.domElement);
		}
	}

	return {
		setMeshes,
		clear,
		requestRender,
		resize,
		fitCamera,
		setCameraPreset,
		setTurntable,
		rotateModelY,
		setRenderMode,
		setWireframe,
		setLightingPreset,
		setMeshVisibility,
		applyToggles,
		captureSnapshot,
		destroy
	};
}
