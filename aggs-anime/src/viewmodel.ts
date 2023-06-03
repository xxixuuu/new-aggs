import EventEmitter from "eventemitter3";

type EventTypes = "playing" | "progress-rate" | "speed";
export class PlayerControlViewModel {
  private ee = new EventEmitter<EventTypes>();
  private playing = false;
  private progressRate = 0;
  private speed = 0.01;
  private prevTickTime = Date.now();

  constructor(
    private readonly minTime: number,
    private readonly maxTime: number
  ) {}

  get isFinished() {
    return this.progressRate === 1;
  }

  play() {
    if (this.isFinished) this.setProgressRate(0);

    this.prevTickTime = Date.now();
    this.playing = true;
    this.ee.emit("playing", this.playing);
  }
  togglePlaying() {
    if (!this.playing && this.isFinished) this.setProgressRate(0);

    this.prevTickTime = Date.now();
    this.playing = !this.playing;
    this.ee.emit("playing", this.playing);
  }
  advanceProgressRate(r: number) {
    this.progressRate += r;
    this.ee.emit("progress-rate", this.progressRate);
  }
  setSpeed(speed: number) {
    this.speed = speed;
    this.ee.emit("speed", this.speed);
  }
  tickTime() {
    if (!this.playing) return;

    const now = Date.now();
    const elapsed = now - this.prevTickTime;
    this.prevTickTime = now;

    const r = (elapsed / (this.maxTime - this.minTime) / 1000) * this.speed;
    this.setProgressRate(Math.min(1, this.progressRate + r));

    if (this.isFinished) {
      this.playing = false;
      this.ee.emit("playing", this.playing);
    }
  }
  setProgressRate(r: number) {
    this.progressRate = r;
    this.ee.emit("progress-rate", this.progressRate);
  }

  subscribePlaying(subscriber: (playing: boolean) => void) {
    this.ee.on("playing", subscriber);
  }
  subscribeProgressRate(subscriber: (progressRate: number) => void) {
    this.ee.on("progress-rate", subscriber);
  }
  subscribeTime(subscriber: (time: number, progressRate: number) => void) {
    this.ee.on("progress-rate", (rate) => {
      const time = this.minTime + (this.maxTime - this.minTime) * rate;
      subscriber(time, rate);
    });
  }

  deinit() {
    this.ee.removeAllListeners();
  }
}
