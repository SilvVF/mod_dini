import {
  SortBy,
  type CharacterFolder,
  type ModSaved,
  type Subfolder,
} from "#lib/bindings/changeme/index.ts";

const yixuanMod: ModSaved = {
  id: "gamebanana:599967",
  state: 4,
  submission: {
    provider: "gamebanana",
    submissionType: "Mod",
    submissionId: "599967",
    slug: "599967",
    name: "Yixuan Divine Bird",
    author: "DaisyMeow",
    category: "37289",
    thumbnailUrl:
      "https://images.gamebanana.com/img/ss/mods/220-90_6849b6804c047.jpg",
    profileUrl: "https://gamebanana.com/mods/599967",
    description: "",
    hero: "https://images.gamebanana.com/img/ss/mods/800-90_6849b6804c047.jpg",
    isAudio: false,
    isMap: false,
    isNsfw: true,
    isObsolete: false,
    isTombstoned: false,
    isHydrated: true,
    hasFiles: true,
    viewCount: 369240,
    likes: 4286,
    remoteAddedAt: 1749661431,
    remoteUpdatedAt: 1763017818,
    filesUpdatedAt: 0,
    lastSeenSnapshot: null,
    images: "[]",
    audioUrl: null,
    tags: "[]",
    developmentState: null,
    completionPercentage: null,
    authorRemoteId: "3809580",
  },
  submission_files: [],
  tree: {
    files: [],
    total_files: 0,
    has_multiple_files: false,
  },
};

export function createMod(
  id: string,
  name: string,
  author: string,
  thumbnailUrl: string,
  category: string,
): ModSaved {
  return {
    ...yixuanMod,
    id: `gamebanana:${id}`,
    submission: {
      ...yixuanMod.submission,
      submissionId: id,
      slug: id,
      name,
      author,
      category,
      thumbnailUrl,
      profileUrl: `https://gamebanana.com/mods/${id}`,
    },
  };
}

export function createCharacterFolder(
  name: string,
  icon: string,
  gamebananaId: string,
): CharacterFolder {
  return {
    name,
    icon,
    sources: {
      gamebanana: gamebananaId,
      local: name,
    },
  };
}

export function createSubfolder(
  id: string,
  name: string,
  icon: string,
  gamebananaId: string,
): Subfolder {
  return {
    id,
    name,
    custom: true,
    config: {
      visible: true,
      pinned: false,
      collapsed: false,
      sort: { by: SortBy.SortByName, ascending: true },
    },
    sources: {
      gamebanana: gamebananaId,
      local: name,
    },
  };
}

const THUMB_DEFAULT =
  "https://images.gamebanana.com/img/ss/mods/220-90_6849b6804c047.jpg";

export const testModRecord = () => {
  let obj: { [k: string]: ModSaved } = {};
  testMods.forEach((m) => {
    obj[m.id] = m;
  });
  return obj;
};

export const testMods: ModSaved[] = [
  // Yixuan (37289)
  yixuanMod,
  createMod(
    "599968",
    "Yixuan Celestial Outfit",
    "Mika",
    THUMB_DEFAULT,
    "37289",
  ),
  createMod(
    "599969",
    "Yixuan Alternate Hair",
    "DaisyMeow",
    THUMB_DEFAULT,
    "37289",
  ),
  createMod(
    "600022",
    "Yixuan Alternate Battle Outfit",
    "Mika",
    THUMB_DEFAULT,
    "37289",
  ),
  createMod(
    "600023",
    "Yixuan Casual Outfit",
    "DaisyMeow",
    THUMB_DEFAULT,
    "37289",
  ),
  createMod("600024", "Yixuan Long Hair", "Kuro", THUMB_DEFAULT, "37289"),

  // Hoshimi Miyabi (30579)
  createMod(
    "600001",
    "Miyabi Sakura Sky",
    "Kuro",
    "https://images.gamebanana.com/img/ss/mods/680535295ca66.jpg",
    "30579",
  ),
  createMod(
    "600002",
    "Miyabi Alternate Outfit",
    "Mika",
    "https://images.gamebanana.com/img/ss/mods/680535295ca66.jpg",
    "30579",
  ),

  // Jane Doe (30580)
  createMod(
    "600003",
    "Jane Doe Echidna",
    "DaisyMeow",
    "https://images.gamebanana.com/img/ss/mods/68224a6c0eeb2.jpg",
    "30580",
  ),
  createMod(
    "600011",
    "Jane Doe Echidna",
    "DaisyMeow",
    "https://images.gamebanana.com/img/ss/mods/68224a6c0eeb2.jpg",
    "30580",
  ),
  createMod(
    "600012",
    "Jane Doe Streetwear",
    "Mika",
    "https://images.gamebanana.com/img/ss/mods/68224a6c0eeb2.jpg",
    "30580",
  ),
  createMod(
    "600013",
    "Jane Doe Alternate Colors",
    "Kuro",
    "https://images.gamebanana.com/img/ss/mods/68224a6c0eeb2.jpg",
    "30580",
  ),

  // Ellen Joe (30341)
  createMod(
    "600004",
    "Ellen Casual Outfit",
    "ExampleAuthor",
    THUMB_DEFAULT,
    "30341",
  ),
  createMod(
    "600005",
    "Ellen Alternate Colors",
    "ExampleAuthor",
    THUMB_DEFAULT,
    "30341",
  ),

  // Von Lycaon (30309)
  createMod(
    "600006",
    "Lycaon Midnight Elegance",
    "DaisyMeow",
    THUMB_DEFAULT,
    "30309",
  ),
  createMod("600007", "Lycaon Formal Attire", "Mika", THUMB_DEFAULT, "30309"),
  createMod("600008", "Lycaon Alternate Hair", "Kuro", THUMB_DEFAULT, "30309"),

  // Evelyn Chevalier (35123)
  createMod(
    "600009",
    "Evelyn Short Formal Dress",
    "DaisyMeow",
    "https://images.gamebanana.com/img/ss/mods/67c050674874a.jpg",
    "35123",
  ),
  createMod(
    "600010",
    "Evelyn Alternate Outfit",
    "Mika",
    "https://images.gamebanana.com/img/ss/mods/67c050674874a.jpg",
    "35123",
  ),

  // Trigger (36133)
  createMod(
    "600014",
    "Trigger Cyber Bunny",
    "DaisyMeow",
    "https://images.gamebanana.com/img/ss/mods/67f2557671afc.jpg",
    "36133",
  ),
  createMod(
    "600015",
    "Trigger Tactical Outfit",
    "Mika",
    "https://images.gamebanana.com/img/ss/mods/67f2557671afc.jpg",
    "36133",
  ),

  // Astra Yao (34740)
  createMod("600016", "Astra Yao Stage Outfit", "Kuro", THUMB_DEFAULT, "34740"),
  createMod(
    "600017",
    "Astra Yao Alternate Dress",
    "Mika",
    THUMB_DEFAULT,
    "34740",
  ),

  // Nicole Demara (30346)
  createMod(
    "600018",
    "Nicole Casual Outfit",
    "ExampleAuthor",
    THUMB_DEFAULT,
    "30346",
  ),
  createMod("600019", "Nicole Alternate Hair", "Mika", THUMB_DEFAULT, "30346"),

  // Soldier 0 Anby (35835)
  createMod(
    "600020",
    "Soldier 0 Anby Sinful Bunny",
    "DaisyMeow",
    "https://images.gamebanana.com/img/ss/mods/67d9515f53e24.jpg",
    "35835",
  ),
  createMod(
    "600021",
    "Soldier 0 Anby Alternate Outfit",
    "Kuro",
    "https://images.gamebanana.com/img/ss/mods/67d9515f53e24.jpg",
    "35835",
  ),
];
