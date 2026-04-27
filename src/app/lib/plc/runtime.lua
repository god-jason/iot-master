
TON = {}
TON.__index = TON

function TON:new()
  return setmetatable({t=0,Q=false}, self)
end

function TON:update(dt, IN, PT)
  if IN then
    self.t = self.t + dt
    self.Q = self.t >= PT
  else
    self.t = 0
    self.Q = false
  end
  return self.Q
end


TOF = {}
TOF.__index = TOF

function TOF:new()
  return setmetatable({t=0,Q=true}, self)
end

function TOF:update(dt, IN, PT)
  if IN then
    self.t = 0
    self.Q = true
  else
    self.t = self.t + dt
    if self.t >= PT then
      self.Q = false
    end
  end
  return self.Q
end


TP = {}
TP.__index = TP

function TP:new()
  return setmetatable({t=0,Q=false,trig=false}, self)
end

function TP:update(dt, IN, PT)
  if IN and not self.trig then
    self.trig = true
    self.Q = true
    self.t = 0
  end

  if self.trig then
    self.t = self.t + dt
    if self.t >= PT then
      self.Q = false
      self.trig = false
    end
  end

  return self.Q
end
