import { Point } from "./geometry";
import { Network, Node } from "./network";
import "./style.css";
import { compare, formatSec, groupBy } from "./util";
import { PlayerControlViewModel } from "./viewmodel";
import { Visualization } from "./visualization";
import sample1 from "./samples/1.json";
import sample2 from "./samples/2.json";
import sample3 from "./samples/3.json";

const svgPlay = `
<svg width="16" height="16" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
  <path d="M4 2L4 14L12 8Z" fill="currentColor"/>
</svg>`;

const svgStop = `
<svg width="16" height="16" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
  <path d="M2 2L2 14L6 14L6 2Z" fill="currentColor"/>
  <path d="M9 2L9 14L13 14L13 2Z" fill="currentColor"/>
</svg>`;

function initCanvas(
  canvas: HTMLCanvasElement,
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number
) {
  const dpr = window.devicePixelRatio;

  canvas.setAttribute("width", `${Math.floor(width * dpr)}`);
  canvas.setAttribute("height", `${Math.floor(height * dpr)}`);

  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;

  ctx.scale(dpr, dpr);
}

function coordinateLayout(width: number, height: number, network: Network) {
  const { nodes } = network;

  let depth = 1;
  let index = 1;
  const depthOf = new Map<string, number>();
  const indexOf = new Map<string, number>();
  function traverse(node: Node) {
    if (depthOf.has(node.id)) return;

    depthOf.set(node.id, depth);
    indexOf.set(node.id, index);

    depth++;
    for (const n of network.connectedNodesOf(node)) {
      index++;
      traverse(n);
    }
    depth--;
  }

  function compareIndex(a: Node, b: Node) {
    return compare(indexOf.get(a.id)!, indexOf.get(b.id)!);
  }

  const roots = network.nodes.filter((_) => _.role === "verifier");
  for (const root of roots) traverse(root);

  const maxDepth = Math.max(...depthOf.values());

  const nodesByDepth = groupBy(nodes, (_) => depthOf.get(_.id)!);
  for (const depth of nodesByDepth.keys()) {
    const nodes = nodesByDepth.get(depth);
    if (nodes == null) continue;

    const w = width / (maxDepth + 1);
    const x = (maxDepth - depth + 1) * w;
    const ns = [...nodes];
    ns.sort(compareIndex);
    const len = ns.length;
    const h = height / (len + 1);
    for (let i = 0; i < len; i++) {
      const n = ns[i];
      const y = (i + 1) * h;
      n.p.x = x;
      n.p.y = y;
    }
  }
}

async function startAnimation(
  network: Network,
  canvas: HTMLCanvasElement,
  vm: PlayerControlViewModel
) {
  const ctx = canvas.getContext("2d");
  if (ctx == null) return;

  const width = 800;
  const height = 600;

  initCanvas(canvas, ctx, width, height);

  coordinateLayout(width, height, network);

  const visualization = new Visualization(width, height, network);

  canvas.addEventListener("pointerdown", function (e) {
    visualization.click(new Point(e.offsetX, e.offsetY));
  });

  vm.subscribeProgressRate((r) => {
    network.setProgressRate(r);
  });

  vm.play();

  const tick = () => {
    vm.tickTime();

    visualization.render(ctx);
    window.requestAnimationFrame(tick);
  };
  tick();
}

function queryEnsure<E extends Element>(q: string): E {
  const el = document.querySelector<E>(q);
  if (el == null) throw new Error(`element ${q} is not found`);

  return el;
}

function initUI(vm: PlayerControlViewModel): () => void {
  const seekBarMax = 10000;

  function onSpeedChange(this: HTMLSelectElement) {
    vm.setSpeed(Number(this.value));
  }
  function onProgressRateInput(this: HTMLInputElement) {
    const v = Number(this.value);
    vm.setProgressRate(v / seekBarMax);
  }

  function onPlayStopButtonClick() {
    vm.togglePlaying();
  }

  const inputSpeed = queryEnsure<HTMLSelectElement>("#input-speed");
  inputSpeed.addEventListener("change", onSpeedChange);

  const seekBar = queryEnsure<HTMLInputElement>("#seek-bar");
  seekBar.setAttribute("max", seekBarMax.toFixed());
  seekBar.addEventListener("input", onProgressRateInput);
  const timeText = queryEnsure<HTMLSpanElement>("#time");
  vm.subscribeProgressRate((r) => {
    seekBar.value = (r * 1000).toString();
  });
  vm.subscribeTime((time, rate) => {
    seekBar.value = (rate * seekBarMax).toString();
    timeText.textContent = formatSec(time);
  });

  const playStopButton = queryEnsure<HTMLButtonElement>("#play-stop-button");
  playStopButton.addEventListener("click", onPlayStopButtonClick);
  vm.subscribePlaying((playing) => {
    if (playing) {
      playStopButton.innerHTML = svgStop;
    } else {
      playStopButton.innerHTML = svgPlay;
    }
  });

  return function deinit() {
    inputSpeed.removeEventListener("change", onSpeedChange);
    seekBar.removeEventListener("input", onProgressRateInput);
    playStopButton.removeEventListener("click", onPlayStopButtonClick);
  };
}

async function main() {
  const canvas = queryEnsure<HTMLCanvasElement>("#main-canvas");
  const dialog = queryEnsure<HTMLDialogElement>("#input-dialog");
  const inputTextarea = queryEnsure<HTMLTextAreaElement>("#input-json");

  queryEnsure<HTMLButtonElement>("#edit-button").addEventListener(
    "click",
    function () {
      dialog.showModal();
    }
  );
  queryEnsure<HTMLButtonElement>("#close-dialog-button").addEventListener(
    "click",
    function () {
      dialog.close();
    }
  );

  queryEnsure("#sample1-button").addEventListener("click", function () {
    inputTextarea.value = JSON.stringify(sample1, null, 2);
  });
  queryEnsure("#sample2-button").addEventListener("click", function () {
    inputTextarea.value = JSON.stringify(sample2, null, 2);
  });
  queryEnsure("#sample3-button").addEventListener("click", function () {
    inputTextarea.value = JSON.stringify(sample3, null, 2);
  });

  dialog.showModal();

  let vm: PlayerControlViewModel | null = null;
  let deinit: (() => void) | null = null;

  queryEnsure("#start-button").addEventListener("click", () => {
    queryEnsure<HTMLDialogElement>("#input-dialog").close();

    if (deinit) deinit();
    if (vm) vm.deinit();

    const jsonInput = inputTextarea;
    const network = Network.fromInput(JSON.parse(jsonInput.value));

    vm = new PlayerControlViewModel(network.minTime, network.maxTime);
    deinit = initUI(vm);

    startAnimation(network, canvas, vm);
  });
}

main().catch((err) => {
  console.error(err);
});
