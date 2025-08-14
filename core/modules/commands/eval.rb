# frozen_string_literal: true

module YourSaySan
  module Commands
    # Evalコマンドモジュール
    module Eval
      extend Discordrb::Commands::CommandContainer

      command :eval do |event, *code|
        return unless event.user.id == YourSaySan::CONFIG.bot.owner

        begin
          result = eval code.join(' ')
          event.respond result.to_s
        rescue StandardError => e
          event.respond "Error: #{e.message}"
        end
        nil
      end
    end
  end
end
