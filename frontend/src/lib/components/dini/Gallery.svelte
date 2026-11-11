<script lang="ts">
  import {
    LightboxGallery,
    GalleryImage,
    GalleryThumbnail,
  } from "svelte-lightbox";

  import NSFWBlur from "./NSFWBlur.svelte";
  import * as Card from "../ui/card";
  import Button from "../ui/button/button.svelte";
  import { Eye, EyeOff } from "@lucide/svelte";

  interface NSFWSettings {
    disableBlur: boolean;
    blurStrength: number;
  }

  interface Props {
    images: string[];
    shouldBlur?: boolean;
    nsfwSettings: NSFWSettings;
    onNSFWToggle?: (visible: boolean) => void;
  }

  let {
    images,
    shouldBlur = false,
    nsfwSettings,
    onNSFWToggle,
  }: Props = $props();

  let galleryVisible = $state(false);
  let galleryResetVersion = $state(0);

  const hasBlurredNsfwImages = $derived(
    shouldBlur && !nsfwSettings.disableBlur,
  );

  function toggleGalleryVisibility() {
    galleryVisible = !galleryVisible;

    if (!galleryVisible) {
      galleryResetVersion++;
    }

    onNSFWToggle?.(galleryVisible);
  }
</script>

{#if images.length}
  <Card.Root class="shadow-none contain-[layout_style_paint]">
    <Card.Header class="flex flex-row items-center justify-between gap-3">
      <Card.Title>Gallery</Card.Title>

      {#if hasBlurredNsfwImages}
        <Button
          class="gap-1 text-xs"
          size="sm"
          variant="secondary"
          onclick={toggleGalleryVisibility}
        >
          {#if galleryVisible}
            <EyeOff class="h-3 w-3" />
            Hide NSFW
          {:else}
            <Eye class="h-3 w-3" />
            Show NSFW
          {/if}
        </Button>
      {/if}
    </Card.Header>

    <Card.Content>
      <LightboxGallery imagePreset='fullscreen' keepBodyScroll={true}>
        <svelte:fragment slot="thumbnail">
          <div class="grid grid-cols-3 gap-2 sm:grid-cols-4">
            {#each images as image, index (`${image}-${galleryResetVersion}`)}
              <GalleryThumbnail id={index}>
                <button
                  type="button"
                  class="group relative overflow-hidden rounded-md border border-border bg-muted text-left ring-offset-background transition outline-none focus-visible:ring-2 focus-visible:ring-ring"
                >
                  <NSFWBlur
                    blurStrength={nsfwSettings.blurStrength}
                    class="aspect-video w-full"
                    disableBlur={nsfwSettings.disableBlur}
                    isNSFW={hasBlurredNsfwImages && !galleryVisible}
                    onToggleVisibility={onNSFWToggle}
                  >
                    <img
                      src={image}
                      alt={`Screenshot ${index + 1}`}
                      loading="lazy"
                      decoding="async"
                      width="400"
                      height="225"
                      class="aspect-video w-full object-contain"
                    />
                  </NSFWBlur>

                  <span
                    class="pointer-events-none absolute inset-0 rounded-md ring-1 ring-black/5 ring-inset group-hover:bg-black/5"
                  ></span>
                </button>
              </GalleryThumbnail>
            {/each}
          </div>
        </svelte:fragment>

        {#each images as image, index}
          <GalleryImage title={``} description={``}>
            <img src={image} alt={`Screenshot ${index + 1}`} />
          </GalleryImage>
        {/each}
      </LightboxGallery>
    </Card.Content>
  </Card.Root>
{/if}