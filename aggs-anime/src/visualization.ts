import { Point, Vector } from "./geometry";
import { Network, Node } from "./network";
import { darken, desaturate, lighten } from "./color-util";
import { formatSec } from "./util";

const nodeRadius = 20;
const fontSize = 16;

const nodeColorMap = {
  signer: lighten("#9b59b6", 0.15),
  aggregator: lighten("#3498db", 0.15),
  verifier: lighten("#2ecc71", 0.15),
} as const;

function renderPolygon(
  ctx: CanvasRenderingContext2D,
  p: Point,
  r: number,
  v: number
) {
  const base = new Vector(0, -r);
  const vertices: Point[] = [];
  for (let i = 0; i < v; i++) {
    const vec = base.rotate((2 * Math.PI * i) / v);
    vertices.push(p.add(vec));
  }

  ctx.beginPath();
  ctx.moveTo(vertices[v - 1].x, vertices[v - 1].y);
  for (const v of vertices) ctx.lineTo(v.x, v.y);
}

function renderNode(ctx: CanvasRenderingContext2D, node: Node) {
  const vertexCount = {
    signer: 3,
    aggregator: 4,
    verifier: 5,
  }[node.role];
  renderPolygon(ctx, node.p, nodeRadius, vertexCount);
}

export class Visualization {
  private selected:
    | { type: "node"; id: string }
    | { type: "message"; id: string }
    | null = null;
  constructor(
    public readonly width: number,
    public readonly height: number,
    public readonly network: Network
  ) {}

  private innerColorOf(node: Node, selected: boolean) {
    let color = nodeColorMap[node.role];
    if (selected) {
      color = lighten(color, 0.1);
    } else {
      color = lighten(color, 0.25);
    }
    if (this.network.isMalicious(node)) {
      color = desaturate(color, 0.75);
      color = darken(color, 0.2);
    }
    return color;
  }

  private borderColorOf(node: Node, selected: boolean) {
    let color = nodeColorMap[node.role];
    if (selected) {
      color = darken(color, 0.15);
    }
    if (this.network.isMalicious(node)) {
      color = desaturate(color, 0.75);
      color = darken(color, 0.2);
    }
    return color;
  }

  render(ctx: CanvasRenderingContext2D) {
    ctx.fillStyle = "#f8f8f8";
    ctx.fillRect(0, 0, 800, 600);

    for (const edge of this.network.edges) {
      ctx.strokeStyle = "#aaa";
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(edge.from.p.x, edge.from.p.y);
      ctx.lineTo(edge.to.p.x, edge.to.p.y);
      ctx.stroke();
    }

    for (const node of this.network.nodes) {
      const selected =
        (this.selected &&
          this.selected.type === "node" &&
          this.selected.id === node.id) ||
        false;

      ctx.fillStyle = this.innerColorOf(node, selected);
      renderNode(ctx, node);
      ctx.fill();

      ctx.strokeStyle = this.borderColorOf(node, selected);
      ctx.lineWidth = 2;
      renderNode(ctx, node);
      ctx.stroke();
    }

    for (const message of this.network.messages) {
      const r = message.packet.size / 2;
      const p = message.p;
      ctx.fillStyle = message.packet.color;
      ctx.beginPath();
      ctx.arc(p.x, p.y, r, 0, 2 * Math.PI);
      ctx.fill();
    }

    const selectedNode =
      this.selected &&
      this.selected.type === "node" &&
      this.network.nodes.find((_) => _.id === this.selected!.id);
    if (selectedNode) {
      const n = selectedNode;
      const w = 200;
      const h = 100;

      const x = n.p.x < this.width / 2 ? n.p.x + 20 : n.p.x - 20 - w;
      const y = n.p.y < this.height / 2 ? n.p.y + 20 : n.p.y - 20 - h;

      ctx.fillStyle = "#fff";
      ctx.fillRect(x, y, w, h);

      ctx.lineWidth = 1;
      ctx.strokeStyle = "#666";
      ctx.strokeRect(x, y, w, h);

      ctx.fillStyle = "#666";
      ctx.font = `${fontSize}px sans-serif`;
      const p = 4;
      ctx.fillText(`id: ${n.id}`, x + p, y + fontSize + p, w);
      ctx.fillText(`role: ${n.role}`, x + p, y + (fontSize + p) * 2, w);
    }

    const selectedMessage =
      this.selected &&
      this.selected.type === "message" &&
      this.network.messages.find((_) => _.packet.id === this.selected!.id);
    if (selectedMessage) {
      const m = selectedMessage;
      const w = 200;
      const h = 100;

      const x = m.p.x < this.width / 2 ? m.p.x + 20 : m.p.x - 20 - w;
      const y = m.p.y < this.height / 2 ? m.p.y + 20 : m.p.y - 20 - h;

      ctx.fillStyle = "#fff";
      ctx.fillRect(x, y, w, h);

      ctx.lineWidth = 1;
      ctx.strokeStyle = "#666";
      ctx.strokeRect(x, y, w, h);

      ctx.fillStyle = "#666";
      ctx.font = `${fontSize}px sans-serif`;
      const p = 4;
      ctx.fillText(`from: ${m.from.id}`, x + p, y + fontSize + p, w);
      ctx.fillText(`to: ${m.to.id}`, x + p, y + (fontSize + p) * 2, w);
      ctx.fillText(
        `sent: ${formatSec(m.packet.sentAt)}`,
        x + p,
        y + (fontSize + p) * 3,
        w
      );
      ctx.fillText(
        `received: ${formatSec(m.packet.receivedAt)}`,
        x + p,
        y + (fontSize + p) * 4,
        w
      );
    }
  }

  click(p: Point) {
    for (const message of this.network.messages) {
      const r = message.packet.size / 2;
      if (message.p.diff(p).abs < r) {
        this.selected = { type: "message", id: message.packet.id };
        return;
      }
    }

    for (const node of this.network.nodes) {
      if (node.p.diff(p).abs < nodeRadius) {
        this.selected = { type: "node", id: node.id };
        return;
      }
    }

    this.selected = null;
  }
}
