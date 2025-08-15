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
            self.say(event.voice, message, voicevox)
          else
            text_channels.delete(event.channel.id)
            nil
          end
        end
      end

      def self.transform_message(bot, server, text, config)
        message = text
        message = message.gsub(/<@([0-9]{18})>/) do
          "@#{bot.member(server, Regexp.last_match(1)).display_name}"
        end
        message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
        message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
        message
      end

      def self.say(voice_chat, message, voicevox)
        Tempfile.create do |tempfile|
          sound = voicevox.speak(message)
          tempfile.write(sound)
          voice_chat.play_file(tempfile.path)
        end
      end
    end
  end
end

