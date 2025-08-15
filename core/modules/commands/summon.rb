# frozen_string_literal: true

module YourSaySan
  module Commands
    # Summonコマンドモジュール
    module Summon
      extend Discordrb::Commands::CommandContainer

      command :summon do |event|
        if event.author.voice_channel
          YourSaySan::BOT.voice_connect(event.author.voice_channel)
          YourSaySan.text_channels << event.channel.id unless YourSaySan.text_channels.include?(event.channel.id)
          'ボイスチャットに参加しました。'
        else
          'ボイスチャットに参加してください。'
        end
      end
    end
  end
end
