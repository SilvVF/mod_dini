import { Events } from "@wailsio/runtime";
import { ModService } from "../bindings/changeme";

export const appState = $state({
  activeTool: "zzmi",
});

const validIds = ["zzmi", "wwmi", "srmi", "gimi"];

export const tools = [
  {
    id: "gimi",
    src: "https://images.gamebanana.com/img/ico/games/686d7fdcbeb5a.png",
    providers: ["gamebanana"],
  },
  {
    id: "wwmi",
    src: "https://images.gamebanana.com/img/ico/games/686bf76048ba2.png",
    providers: ["gamebanana", "huihui"],
  },
  {
    id: "zzmi",
    src: "https://images.gamebanana.com/img/ico/games/657d72ea51567.png",
    providers: ["gamebanana", "huihui"],
  },
  {
    id: "srmi",
    src: "https://images.gamebanana.com/img/ico/games/686d7fe451b81.png",
    providers: ["gamebanana", "huihui"],
  },
];

export function setActiveTool(toolId: string) {
  if (validIds.includes(toolId)) {
    appState.activeTool = toolId;
    ModService.SetActiveTool(toolId);
  }
}

export function toolName(toolId: string): string {
  switch (toolId) {
    case "zzmi":
      return "Zenless Zone Zero";
    case "srmi":
      return "Honkai Star Rail";
    case "wwmi":
      return "Wuthering Waves";
    case "gimi":
      return "Genshin Impact";
  }
  return "";
}

export function getRouteForProvider(provider: string): string | undefined {
  const conf = tools.find((t) => t.id === appState.activeTool)!;
  if (conf.providers.some((p) => p === provider)) {
    switch (provider) {
      case "huihui":
        return `/source/huihui/${currentHuiHuiCategory()}`;
      case "gamebanana":
        return `/source/gamebanana/${currentGamebannaCategory()}`;
    }
  }

  return undefined;
}

export function currentHuiHuiCategory() {
  switch (appState.activeTool) {
    case "zzmi":
      return "zzz";
    case "srmi":
      return "starrail";
    case "wwmi":
      return "wuwa";
    case "gimi":
      return "";
  }
}

export function currentGamebannaCategory() {
  switch (appState.activeTool) {
    case "zzmi":
      return "30305";
    case "srmi":
      return "22832";
    case "wwmi":
      return "29524";
    case "gimi":
      return "18140";
  }
}

export function toolBanner(toolId: string): string {
  switch (toolId) {
    case "zzmi":
      return "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/66868c3874664.jpg";
    case "srmi":
      return "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/6876c3e492e8d.jpg";
    case "wwmi":
      return "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/6654974ecec33.jpg";
    case "gimi":
      return "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/68c8f553e9533.png";
  }
  return "";
}
