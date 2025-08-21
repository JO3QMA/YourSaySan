# frozen_string_literal: true

module YourSaySan
  module Commands
    # SpeakerListコマンドモジュール
    module SpeakerList
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: '利用可能な話者の一覧を表示します。',
        usage: '`/speaker_list [page]` でページ番号を指定して話者一覧を表示します。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:speaker_list, COMMAND_INFO[:desc]) do |cmd|
          cmd.integer(:page, 'ページ番号（1から開始）', required: false)
        end
      end

      application_command :speaker_list do |event|
        speaker_manager = YourSaySan.speaker_manager
        
        if speaker_manager.nil?
          event.respond(content: "話者マネージャーが初期化されていません。VoiceVoxの設定を確認してください。", ephemeral: true)
          return
        end
        
        speakers = speaker_manager.get_available_speakers

        if speakers.nil?
          event.respond(content: "VoiceVoxのAPIに接続できません。VoiceVoxが起動しているか確認してください。", ephemeral: true)
          return
        end

        # 現在のユーザーの話者設定を取得
        current_speaker_id = speaker_manager.get_speaker(event.user.id)
        current_speaker_name = speakers[current_speaker_id] || "不明な話者"

        # ページネーション設定
        page = event.options['page'] || 1
        items_per_page = 20
        total_pages = (speakers.length.to_f / items_per_page).ceil
        page = [page, total_pages].min # 最大ページ数を超えないように
        page = [page, 1].max # 最小ページ数未満にならないように

        # 表示する話者の範囲を計算
        start_index = (page - 1) * items_per_page
        end_index = [start_index + items_per_page - 1, speakers.length - 1].min
        page_speakers = speakers.to_a[start_index..end_index]

        # 話者一覧を作成
        speaker_list = page_speakers.map do |id, name|
          marker = (id == current_speaker_id) ? "▶ " : "  "
          "#{marker}#{id}: #{name}"
        end.join("\n")

        max_id = speakers.keys.max
        embed = Discordrb::Webhooks::Embed.new(
          title: "利用可能な話者一覧",
          description: "現在の設定: **#{current_speaker_id}: #{current_speaker_name}**\n\n```\n#{speaker_list}\n```\n\n`/speaker 話者ID（0-#{max_id}）` で話者を設定できます。",
          color: 0x00ff00,
          footer: { text: "ページ #{page}/#{total_pages}（全#{speakers.length}件）" }
        )

        # ページネーション情報を追加
        if total_pages > 1
          embed.add_field(
            name: "ページネーション",
            value: "`/speaker_list page:#{page}` でページを指定できます。\n" +
                   "前のページ: #{page > 1 ? page - 1 : 'なし'}\n" +
                   "次のページ: #{page < total_pages ? page + 1 : 'なし'}",
            inline: false
          )
        end

        event.respond(embeds: [embed], ephemeral: true)
      end
    end
  end
end
