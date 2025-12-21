module Packing
  class Worker
    include WithAttributes
    attributes :id, :capacity, :power

    def can_take?(job)
      capacity > 0 && power >= job.power
    end

    def assign!(job)
      @capacity -= 1
      @power -= job.power
      job.worker = id
    end
  end
end
