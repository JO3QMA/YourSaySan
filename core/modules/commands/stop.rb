# frozen_string_literal: true

# Command Module
module YourSaySan
  module Commands
    # Stopコマンドモジュール
    module Stop
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: '読み上げを中断します。',
        usage: '引数は必要ありません。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:stop, COMMAND_INFO[:desc])
      end

      application_command :stop do |event|
        if YourSaySan::BOT.voice(event.server)
          YourSaySan::BOT.voice(event.server).stop_playing
          event.respond(content: '読み上げを中断しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに接続していません。', ephemeral: true)
        end
      end
    end
  end
end
