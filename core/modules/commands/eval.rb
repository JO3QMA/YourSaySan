# frozen_string_literal: true

# Command Module
module YouSaySan
  module Commands
    # eval command that executes the input content
    module Eval
      extend Discordrb::Commands::CommandContainer
      command(:eval, { aliases: [:e], help_available: false }) do |event, *code|
        break unless event.user.id == BOT::CONFIG.bot.owner

        begin
          eval code.join(' ')
        rescue StandardError
          event.respond('エラーが発生しました。')
          event.respond($!)
        end
      end
    end
  end
end
