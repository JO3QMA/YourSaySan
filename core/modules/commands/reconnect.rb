# frozen_string_literal: true

module YourSaySan
  module Commands
    # Reconnectコマンドモジュール
    module Reconnect
      extend Discordrb::Commands::CommandContainer

      command :reconnect do |event|
        if event.author.voice_channel
          YourSaySan::BOT.voice_destroy(event.server)
          YourSaySan::BOT.voice_connect(event.author.voice_channel)
          '再接続しました。'
        else
          'ボイスチャットに参加してください。'
        end
      end
    end
  end
end


