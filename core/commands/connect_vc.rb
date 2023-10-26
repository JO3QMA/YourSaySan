# frozen_string_literal: true

# Command Module
module YouSaySan
  module Commands
    # Connect to VC command.
    module Summon
      extend Discordrb::Commands::CommandContainer
      command(:summon, { aliases: [:s] }) do |event|
        if event.user.voice_channel
          if !event.voice
            BOT.voice_connect(event.user.voice_channel)
            @text_channels << event.channel.id
            event.respond('こんにちは! 読み上げBOTです!')
            event.respond("`#{BOT::CONFIG.bot.prefix}help` でコマンドの使い方を表示します。")
          else
            event.respond('すでにVCに参加しています。')
          end
        else
          event.respond('VCに参加した状態で使用してください。')
        end
      end
    end

    # Disconnect from VC command.
    module Bye
      extend Discordrb::Commands::CommandContainer
      command(:bye, { aliases: [:b] }) do |event|
        if event.voice
          @text_channels.delete(event.channel.id)
          event.voice.destroy
          event.respond('Bye!')
        end
      end
    end

    # Reconnect to VC command.
    module Reconnect
      extend Discordrb::Commands::CommandContainer
      comand(:reconnect, { aliases: [:re] }) do |event|
        if event.voice
          vc = event.voice.channel
          event.voice.destroy
          BOT.voice_connect(vc)
          event.respond('再接続しました。')
        else
          event.respond('VCに参加していません。')
        end
      end
    end
  end
end
