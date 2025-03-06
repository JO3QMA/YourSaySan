# frozen_string_literal: true

module YourSaySan
  module Commands
    # Evalコマンドモジュール
    module Eval
      extend Discordrb::Commands::CommandContainer

      command :eval do |event, *code|
        if event.user.id == YourSaySan::CONFIG.bot.owner
          begin
            result = eval code.join(' ')
            result.to_s
          rescue StandardError => e
            "Error: #{e}"
          end
        end
        event.respond result.to_s
      end
    end
  end
end
