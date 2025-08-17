# frozen_string_literal: true

module YourSaySan
  module Commands
    # Speakerコマンドモジュール
    module Speaker
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:speaker, '話者を設定します') do |cmd|
          cmd.string(:speaker_id, '話者ID', required: true)
        end
      end

      application_command :speaker do |event|
        speaker_manager = YourSaySan.speaker_manager
        speakers = speaker_manager.get_available_speakers

        if speakers.nil?
          event.respond(content: "VoiceVoxのAPIに接続できません。VoiceVoxが起動しているか確認してください。", ephemeral: true)
          return
        end

        speaker_id = event.options['speaker_id'].to_i

        if speaker_manager.valid_speaker?(speaker_id)
          speaker_manager.set_speaker(event.user.id, speaker_id)
          speaker_name = speakers[speaker_id]
          event.respond(content: "話者を「#{speaker_name}」に設定しました。", ephemeral: true)
        else
          max_id = speakers.keys.max
          event.respond(content: "無効な話者IDです。0-#{max_id}の範囲で指定してください。", ephemeral: true)
        end
      end
      
    end
  end
end
