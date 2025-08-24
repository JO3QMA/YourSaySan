# frozen_string_literal: true

# Command Module
module YourSaySan
  module Events
    module Read
      extend Discordrb::EventContainer

      require 'tempfile'
      require 'uri'

      def self.setup(bot, text_channels, voicevox, config)
        bot.message(in: text_channels) do |event|
          # スラッシュコマンドの場合は読み上げない
          next if event.message.content.start_with?('/')

          if event.voice
            next if event.message.content == '' # 画像など本文がない投稿を弾く

            message = transform_message(bot, event.server, event.message.content, config)
            say(event.voice, message, voicevox, event.user.id)
          else
            text_channels.delete(event.channel.id)
            nil
          end
        end
      end

      def self.transform_message(bot, server, text, config)
        message = text
        message = message.gsub(/<@([0-9]{18})>/) {
          "@#{bot.member(server, Regexp.last_match(1)).display_name}"
        }
        message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
        message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
        message
      end

      def self.say(voice_chat, message, voicevox, user_id)
        # ユーザーごとの話者設定を取得
        speaker_manager = YourSaySan.speaker_manager
        speaker = speaker_manager ? speaker_manager.get_speaker(user_id) : 2

        # VoiceVoxから音声データを取得
        sound_data = voicevox.speak(message, speaker)

        Tempfile.create(['voice', '.wav']) do |tempfile|
          tempfile.binmode
          tempfile.write(sound_data)
          tempfile.flush
          voice_chat.play_file(tempfile.path)
        end
      end
    end
  end
end
