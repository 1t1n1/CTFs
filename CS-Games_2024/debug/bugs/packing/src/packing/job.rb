module Packing
  class Job
    include WithAttributes
    attributes :id, :depends_on, :power, :worker

    def worker=(worker)
      @worker = worker
    end
  end
end
