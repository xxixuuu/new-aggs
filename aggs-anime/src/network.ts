import { Point } from "./geometry";
import { isNotNull } from "./util";

type Role = "signer" | "aggregator" | "verifier";

export class Node {
  constructor(
    public id: string,
    public role: Role,
    public p: Point,
    public maliciousDetectedAt: number | null
  ) {}
}

export class Edge {
  constructor(public from: Node, public to: Node) {}
}

export class Packet {
  id: string;
  constructor(
    public source: Node,
    public destination: Node,
    public sentAt: number,
    public receivedAt: number,
    public size: number,
    public color: string
  ) {
    this.id = [
      this.source.id,
      this.destination.id,
      this.sentAt,
      this.receivedAt,
    ].join(":");
  }
}

export class Message {
  public p: Point;
  constructor(public packet: Packet, public progress: number) {
    const d = this.to.p.diff(this.from.p);
    this.p = this.from.p.add(d.scale(this.progress));
  }

  get from() {
    return this.packet.source;
  }
  get to() {
    return this.packet.destination;
  }
}

export type NetworkInput = {
  nodes: {
    id: string;
    type: Role;
    maliciousDetectedAt?: number;
  }[];
  packets: {
    sourceNodeId: string;
    destinationNodeId: string;
    sentAt: number;
    receivedAt: number;
    size: number;
    color: string;
  }[];
};
export class Network {
  constructor(
    public nodes: Node[],
    public edges: Edge[],
    public packets: Packet[],
    private time: number
  ) {}
  static fromInput(input: NetworkInput): Network {
    const nodes: Node[] = input.nodes.map(
      (_) =>
        new Node(
          _.id,
          _.type,
          Point.origin(),
          _.maliciousDetectedAt == null ? null : _.maliciousDetectedAt / 1000000
        )
    );
    const edgePairSet = new Set<string>();
    const delimiter = "@@";
    for (const packet of input.packets) {
      let a = packet.sourceNodeId;
      let b = packet.destinationNodeId;
      if (b > a) {
        const t = a;
        a = b;
        b = t;
      }

      edgePairSet.add(a + delimiter + b);
    }

    const edges: Edge[] = [];
    for (const edgePair of edgePairSet.values()) {
      const [sourceId, destinationId] = edgePair.split(delimiter, 2);
      const source = nodes.find((_) => _.id === sourceId);
      if (source == null) throw new Error(`node ${sourceId} is not found`);
      const destination = nodes.find((_) => _.id === destinationId);
      if (destination == null)
        throw new Error(`node ${destinationId} is not found`);

      edges.push(new Edge(source, destination));
    }

    const packets: Packet[] = [];
    for (const p of input.packets) {
      const source = nodes.find((_) => _.id === p.sourceNodeId);
      if (source == null)
        throw new Error(`node ${p.sourceNodeId} is not found`);
      const destination = nodes.find((_) => _.id === p.destinationNodeId);
      if (destination == null)
        throw new Error(`node ${p.destinationNodeId} is not found`);

      packets.push(
        new Packet(
          source,
          destination,
          p.sentAt / 1000000,
          p.receivedAt / 1000000,
          p.size,
          p.color
        )
      );
    }

    return new Network(nodes, edges, packets, 0);
  }

  get messages() {
    const items: Message[] = [];
    for (const packet of this.packets) {
      if (packet.sentAt <= this.time && this.time <= packet.receivedAt) {
        const progress =
          (this.time - packet.sentAt) / (packet.receivedAt - packet.sentAt);
        items.push(new Message(packet, progress));
      }
    }

    return items;
  }

  get minTime(): number {
    return Math.min(...this.packets.map((_) => _.sentAt));
  }

  get maxTime(): number {
    return Math.max(...this.packets.map((_) => _.receivedAt));
  }

  isMalicious(node: Node): boolean {
    if (node.maliciousDetectedAt == null) return false;
    return node.maliciousDetectedAt <= this.time;
  }

  connectedNodesOf(node: Node): Node[] {
    return this.edges
      .map((_) =>
        _.from.id === node.id ? _.to : _.to.id === node.id ? _.from : null
      )
      .filter(isNotNull);
  }

  setProgressRate(r: number) {
    this.time = this.minTime + r * (this.maxTime - this.minTime);
  }
}
