# frozen_string_literal: true

require 'discordrb'
require 'config'
require 'yaml'
require 'erb'
require 'logger'
require_relative 'voicevox'
require_relative 'speaker_manager'

# YourSaySanモジュール
module YourSaySan
  # Load config.yml with ERB + YAML so ENV values can be embedded
  CONFIG = begin
    erb = ERB.new(File.read('config.yml'))
    yaml = YAML.safe_load(erb.result, aliases: true)
    Config.load_and_set_settings(yaml)
  end
  BOT = Discordrb::Bot.new(
    token: CONFIG.bot.token,
    client_id: CONFIG.bot.client_id,
    ignore_bots: true
  )

  # Module define

  module Commands; end
  module Events; end

  # Shared states
  @text_channels = []
  @voicevox = VoiceVox.new(CONFIG, Logger.new($stdout)) rescue nil
  @speaker_manager = SpeakerManager.new(CONFIG, Logger.new($stdout), @voicevox) rescue nil

  def self.text_channels
    @text_channels
  end

  def self.voicevox
    @voicevox
  end

  def self.speaker_manager
    @speaker_manager
  end

  # コマンドとイベントを登録する（非同期版）
  def self.load_modules
    puts '[Bot] モジュールの読み込み開始'

    # Bot Commands
    Dir['./core/modules/commands/*.rb'].sort.each do |file|
      puts "[Bot] Load Command: #{file}"
      require file
    end
    Commands.constants.each do |mod|
      BOT.include! Commands.const_get mod
    end

    # Bot Events
    Dir['./core/modules/events/*.rb'].sort.each do |file|
      puts "[Bot] Load Event: #{file}"
      require file
    end
    Events.constants.each do |mod|
      mod_ref = Events.const_get mod
      BOT.include! mod_ref
      # イベントモジュールが setup を提供している場合のみ呼ぶ（テストで差し替え可能）
      mod_ref.setup(BOT, @text_channels, @voicevox, CONFIG) if mod_ref.respond_to?(:setup)
    end

    # 各コマンドモジュールのコマンド登録を非同期で実行
    register_slash_commands_from_modules_async
  end

  # 非同期でコマンド登録を実行する
  def self.register_slash_commands_from_modules_async
    puts '[Bot] コマンドの非同期登録開始'
    start_time = Time.now
    
    # 登録対象のコマンドモジュールを収集
    command_modules = []
    Commands.constants.each do |mod|
      mod_ref = Commands.const_get mod
      if mod_ref.respond_to?(:register_slash_command)
        command_modules << { name: mod, module: mod_ref }
      end
    end

    puts "[Bot] 登録対象コマンド数: #{command_modules.length}"

    # 並行処理でコマンド登録を実行
    threads = command_modules.map do |cmd_info|
      Thread.new do
        thread_start_time = Time.now
        begin
          puts "[Bot] コマンド登録開始: #{cmd_info[:name]}"
          cmd_info[:module].register_slash_command(BOT)
          thread_duration = Time.now - thread_start_time
          puts "[Bot] コマンド登録完了: #{cmd_info[:name]} (#{thread_duration.round(3)}秒)"
        rescue StandardError => e
          thread_duration = Time.now - thread_start_time
          puts "[Bot] コマンド登録エラー (#{cmd_info[:name]}, #{thread_duration.round(3)}秒): #{e.message}"
          puts e.backtrace.join("\n")
        end
      end
    end

    # 全てのスレッドの完了を待機
    threads.each(&:join)
    total_duration = Time.now - start_time
    puts "[Bot] コマンドの非同期登録完了 (総処理時間: #{total_duration.round(3)}秒)"
  end

  # 従来の同期版メソッド（後方互換性のため保持）
  def self.register_slash_commands_from_modules
    puts '[Bot] コマンドの登録開始'
    
    Commands.constants.each do |mod|
      mod_ref = Commands.const_get mod
      if mod_ref.respond_to?(:register_slash_command)
        mod_ref.register_slash_command(BOT)
      end
    end

    puts '[Bot] コマンドの登録完了'
  end

  def self.run
    load_modules
    BOT.run
  rescue StandardError => e
    puts "Bot実行中にエラーが発生しました: #{e.message}"
    puts e.backtrace.join("\n")
  end
end
