
export class TON {
  t = 0;
  Q = false;

  update(dt: number, IN: boolean, PT: number) {
    if (IN) {
      this.t += dt;
      this.Q = this.t >= PT;
    } else {
      this.t = 0;
      this.Q = false;
    }
    return this.Q;
  }
}

export class TOF {
  t = 0;
  Q = true;

  update(dt: number, IN: boolean, PT: number) {
    if (IN) {
      this.Q = true;
      this.t = 0;
    } else {
      this.t += dt;
      if (this.t >= PT) this.Q = false;
    }
    return this.Q;
  }
}

export class TP {
  t = 0;
  Q = false;
  trig = false;

  update(dt: number, IN: boolean, PT: number) {
    if (IN && !this.trig) {
      this.trig = true;
      this.Q = true;
      this.t = 0;
    }

    if (this.trig) {
      this.t += dt;
      if (this.t >= PT) {
        this.Q = false;
        this.trig = false;
      }
    }

    return this.Q;
  }
}

export class PID {
  Kp = 1;
  Ki = 0;
  Kd = 0;

  i = 0;
  last = 0;

  update(dt: number, sp: number, pv: number) {

    const err = sp - pv;

    this.i += err * dt;

    const d = (err - this.last) / dt;

    const out =
      this.Kp * err +
      this.Ki * this.i +
      this.Kd * d;

    this.last = err;

    return out;
  }
}

export class PLC {
  env: any = {
    memory: {},
    fb: {},
    timers: {}
  };

  step(program: Function, dt: number) {
    program(this.env, dt);
  }
}
