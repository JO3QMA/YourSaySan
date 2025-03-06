# frozen_string_literal: true

module YourSaySan
  module Commands
    # Byeコマンドモジュール
    module Bye
      extend Discordrb::Commands::CommandContainer

      command :bye do |event|
        if event.author.voice_channel
          YourSaySan::BOT.voice_destroy(event.server)
          'ボイスチャットから切断しました。'
        else
          'ボイスチャットに参加していません。'
        end
      end
    end
  end
end
