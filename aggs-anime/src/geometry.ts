export class Point {
  constructor(public x: number, public y: number) {}

  static origin() {
    return new Point(0, 0);
  }

  add(v: Vector): Point {
    return new Point(this.x + v.x, this.y + v.y);
  }

  diff(other: Point): Vector {
    return new Vector(this.x - other.x, this.y - other.y);
  }
}

export class Vector {
  constructor(public x: number, public y: number) {}

  scale(s: number): Vector {
    return new Vector(this.x * s, this.y * s);
  }

  rotate(theta: number): Vector {
    const x = this.x * Math.cos(theta) - this.y * Math.sin(theta);
    const y = this.x * Math.sin(theta) + this.y * Math.cos(theta);
    return new Vector(x, y);
  }

  get abs(): number {
    return Math.sqrt(this.x * this.x + this.y * this.y);
  }

  get unit(): Vector {
    const { x, y, abs } = this;
    return new Vector(x / abs, y / abs);
  }
}
