module Packing::Strategies
  class FillFirst < Packing::Strategy
    def schedule(job)
      assigned = false

      spec.workers.each do |worker|
        next unless worker.can_take?(job)

        worker.assign!(job)
        assigned = true
        break
      end

      raise 'Capacity exceeded' unless assigned
    end
  end
end
